package main

import (
    "context"
    "database/sql"
    "net"
    "os"
    "time"

    pb "github.com/MAMUER/Project/api/gen/biometric"
    "github.com/MAMUER/Project/internal/db"
    "github.com/MAMUER/Project/internal/logger"
    "github.com/MAMUER/Project/internal/queue"
    "github.com/google/uuid"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type biometricServer struct {
    pb.UnimplementedBiometricServiceServer
    db          *sql.DB
    log         *logger.Logger
    rabbitQueue *queue.Publisher
}

func (s *biometricServer) AddRecord(ctx context.Context, req *pb.AddRecordRequest) (*pb.AddRecordResponse, error) {
    s.log.Info("AddRecord",
        zap.String("user_id", req.UserId),
        zap.String("metric_type", req.MetricType),
        zap.Float64("value", req.Value),
    )

    id := uuid.New().String()
    timestamp := req.Timestamp.AsTime()

    _, err := s.db.Exec(`
        INSERT INTO biometric_data (id, user_id, metric_type, value, timestamp, device_type)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, id, req.UserId, req.MetricType, req.Value, timestamp, req.DeviceType)
    if err != nil {
        s.log.Error("Failed to insert record", zap.Error(err))
        return nil, status.Error(codes.Internal, "failed to insert record")
    }

    // Отправляем в очередь для ML-классификации
    event := map[string]interface{}{
        "user_id":     req.UserId,
        "metric_type": req.MetricType,
        "value":       req.Value,
        "timestamp":   timestamp,
    }
    if s.rabbitQueue != nil {
        s.rabbitQueue.Publish(ctx, event)
    }

    return &pb.AddRecordResponse{Id: id}, nil
}

func (s *biometricServer) BatchAddRecords(ctx context.Context, req *pb.BatchAddRecordsRequest) (*pb.BatchAddRecordsResponse, error) {
    s.log.Info("BatchAddRecords",
        zap.String("user_id", req.UserId),
        zap.Int("count", len(req.Records)),
    )

    tx, err := s.db.Begin()
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to begin transaction")
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO biometric_data (id, user_id, metric_type, value, timestamp, device_type)
        VALUES ($1, $2, $3, $4, $5, $6)
    `)
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to prepare statement")
    }
    defer stmt.Close()

    for _, record := range req.Records {
        id := uuid.New().String()
        timestamp := record.Timestamp.AsTime()
        _, err := stmt.Exec(id, req.UserId, record.MetricType, record.Value, timestamp, record.DeviceType)
        if err != nil {
            s.log.Error("Failed to insert batch record", zap.Error(err))
            return nil, status.Error(codes.Internal, "failed to insert record")
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, status.Error(codes.Internal, "failed to commit transaction")
    }

    return &pb.BatchAddRecordsResponse{Count: int32(len(req.Records))}, nil
}

func (s *biometricServer) GetRecords(ctx context.Context, req *pb.GetRecordsRequest) (*pb.GetRecordsResponse, error) {
    s.log.Debug("GetRecords",
        zap.String("user_id", req.UserId),
        zap.String("metric_type", req.MetricType),
    )

    from := req.From.AsTime()
    to := req.To.AsTime()

    rows, err := s.db.Query(`
        SELECT id, user_id, metric_type, value, timestamp, device_type, created_at
        FROM biometric_data
        WHERE user_id = $1 AND metric_type = $2 AND timestamp BETWEEN $3 AND $4
        ORDER BY timestamp DESC
        LIMIT $5
    `, req.UserId, req.MetricType, from, to, req.Limit)
    if err != nil {
        s.log.Error("Failed to query records", zap.Error(err))
        return nil, status.Error(codes.Internal, "failed to query records")
    }
    defer rows.Close()

    var records []*pb.BiometricRecord
    for rows.Next() {
        var record pb.BiometricRecord
        var timestamp, createdAt time.Time
        err := rows.Scan(&record.Id, &record.UserId, &record.MetricType, &record.Value,
            &timestamp, &record.DeviceType, &createdAt)
        if err != nil {
            continue
        }
        record.Timestamp = timestamppb.New(timestamp)
        record.CreatedAt = timestamppb.New(createdAt)
        records = append(records, &record)
    }

    return &pb.GetRecordsResponse{Records: records}, nil
}

func (s *biometricServer) GetLatest(ctx context.Context, req *pb.GetLatestRequest) (*pb.BiometricRecord, error) {
    s.log.Debug("GetLatest",
        zap.String("user_id", req.UserId),
        zap.String("metric_type", req.MetricType),
    )

    var record pb.BiometricRecord
    var timestamp, createdAt time.Time

    err := s.db.QueryRow(`
        SELECT id, user_id, metric_type, value, timestamp, device_type, created_at
        FROM biometric_data
        WHERE user_id = $1 AND metric_type = $2
        ORDER BY timestamp DESC
        LIMIT 1
    `, req.UserId, req.MetricType).Scan(
        &record.Id, &record.UserId, &record.MetricType, &record.Value,
        &timestamp, &record.DeviceType, &createdAt,
    )
    if err == sql.ErrNoRows {
        return nil, status.Error(codes.NotFound, "no records found")
    }
    if err != nil {
        s.log.Error("Failed to query latest record", zap.Error(err))
        return nil, status.Error(codes.Internal, "failed to query record")
    }

    record.Timestamp = timestamppb.New(timestamp)
    record.CreatedAt = timestamppb.New(createdAt)

    return &record, nil
}

func main() {
    log := logger.New("biometric-service")
    defer log.Sync()

    port := os.Getenv("BIOMETRIC_SERVICE_PORT")
    if port == "" {
        port = "50052"
    }

    dbCfg := db.Config{
        Host:     os.Getenv("DB_HOST"),
        Port:     os.Getenv("DB_PORT"),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        DBName:   os.Getenv("DB_NAME"),
        SSLMode:  os.Getenv("DB_SSLMODE"),
    }
    database, err := db.NewConnection(dbCfg)
    if err != nil {
        log.Fatal("Failed to connect to database", zap.Error(err))
    }
    defer database.Close()

    rabbitURL := os.Getenv("RABBITMQ_URL")
    queueName := "biometric_events"
    var rabbitQueue *queue.Publisher
    if rabbitURL != "" {
        rabbitQueue, err = queue.NewPublisher(rabbitURL, queueName)
        if err != nil {
            log.Warn("Failed to connect to RabbitMQ", zap.Error(err))
        } else {
            defer rabbitQueue.Close()
            log.Info("RabbitMQ connected", zap.String("queue", queueName))
        }
    }

    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        log.Fatal("Failed to listen", zap.Error(err))
    }

    s := grpc.NewServer()
    pb.RegisterBiometricServiceServer(s, &biometricServer{
        db:          database,
        log:         log,
        rabbitQueue: rabbitQueue,
    })

    log.Info("Biometric service starting", zap.String("port", port))
    if err := s.Serve(lis); err != nil {
        log.Fatal("Failed to serve", zap.Error(err))
    }
}