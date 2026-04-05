package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"os"
	"time"

	pb "github.com/MAMUER/Project/api/gen/training"
	"github.com/MAMUER/Project/internal/db"
	"github.com/MAMUER/Project/internal/logger"
	"github.com/MAMUER/Project/internal/queue"
	"github.com/MAMUER/Project/internal/sanitize"
	"github.com/MAMUER/Project/internal/validator"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type trainingServer struct {
	pb.UnimplementedTrainingServiceServer
	db          *sql.DB
	log         *logger.Logger
	rabbitQueue queue.Publisher // ← ИНТЕРФЕЙС
}

func (s *trainingServer) GeneratePlan(ctx context.Context, req *pb.GeneratePlanRequest) (*pb.GeneratePlanResponse, error) {
	s.log.Info("GeneratePlan",
		zap.String("user_id", req.UserId),
		zap.String("class", req.ClassificationClass),
	)

	// Проверяем отмену контекста
	if err := ctx.Err(); err != nil {
		s.log.Warn("Request cancelled", zap.Error(err))
		return nil, status.Error(codes.Canceled, "request cancelled")
	}

	// Валидация входных данных
	if err := validator.ValidateGeneratePlanRequest(req); err != nil {
		s.log.Warn("Invalid generate plan request", zap.Error(err))
		return nil, err
	}

	// Санитизируем входные данные
	classificationClass := sanitize.String(req.ClassificationClass)

	planID := uuid.New().String()

	planData := map[string]interface{}{
		"name":       "Персонализированная программа",
		"class":      classificationClass,
		"confidence": req.Confidence,
		"weeks":      req.DurationWeeks,
		"schedule":   req.AvailableDays,
		"workouts": []map[string]interface{}{
			{
				"day":       1,
				"type":      "cardio",
				"duration":  30,
				"intensity": "medium",
				"exercises": []string{"бег", "велосипед"},
			},
			{
				"day":       3,
				"type":      "strength",
				"duration":  45,
				"intensity": "high",
				"exercises": []string{"приседания", "отжимания", "тяга"},
			},
			{
				"day":       5,
				"type":      "recovery",
				"duration":  20,
				"intensity": "low",
				"exercises": []string{"растяжка", "йога"},
			},
		},
	}

	planDataJSON, err := json.Marshal(planData)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal plan data")
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, int(req.DurationWeeks)*7)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO training_plans (id, user_id, plan_data, generated_at, start_date, end_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, planID, req.UserId, planDataJSON, time.Now(), startDate, endDate, "active")
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save plan")
	}

	event := map[string]interface{}{
		"event":     "plan_generated",
		"user_id":   req.UserId,
		"plan_id":   planID,
		"class":     classificationClass,
		"timestamp": time.Now(),
	}
	if s.rabbitQueue != nil {
		if err := s.rabbitQueue.Publish(ctx, event); err != nil {
			s.log.Warn("Failed to publish event", zap.Error(err))
		}
	}

	planStruct, err := structpb.NewStruct(planData)
	if err != nil {
		s.log.Error("Failed to create plan struct", zap.Error(err))
		planStruct = &structpb.Struct{}
	}

	return &pb.GeneratePlanResponse{
		PlanId:   planID,
		PlanData: planStruct,
	}, nil
}

func (s *trainingServer) GetPlan(ctx context.Context, req *pb.GetPlanRequest) (*pb.TrainingPlan, error) {
	s.log.Debug("GetPlan", zap.String("plan_id", req.PlanId))

	var plan pb.TrainingPlan
	var planDataJSON []byte
	var generatedAt, startDate, endDate time.Time

	err := s.db.QueryRow(`
		SELECT id, user_id, plan_data, generated_at, start_date, end_date, status
		FROM training_plans
		WHERE id = $1
	`, req.PlanId).Scan(&plan.Id, &plan.UserId, &planDataJSON, &generatedAt, &startDate, &endDate, &plan.Status)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "plan not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "database error")
	}

	var planData map[string]interface{}
	if err := json.Unmarshal(planDataJSON, &planData); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal plan: %v", err)
	}

	planDataStruct, err := structpb.NewStruct(planData)
	if err != nil {
		s.log.Error("Failed to create plan struct", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to process plan data")
	}
	plan.PlanData = planDataStruct
	plan.GeneratedAt = timestamppb.New(generatedAt)
	plan.StartDate = timestamppb.New(startDate)
	plan.EndDate = timestamppb.New(endDate)

	return &plan, nil
}

func (s *trainingServer) ListPlans(ctx context.Context, req *pb.ListPlansRequest) (*pb.ListPlansResponse, error) {
	// Валидация параметров
	if err := validator.ValidateListPlansRequest(req); err != nil {
		s.log.Warn("Invalid list plans request", zap.Error(err))
		return nil, err
	}

	s.log.Debug("ListPlans", zap.String("user_id", req.UserId))

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, plan_data, generated_at, start_date, end_date, status
		FROM training_plans
		WHERE user_id = $1
		ORDER BY generated_at DESC
		LIMIT $2 OFFSET $3
	`, req.UserId, req.PageSize, req.Page*req.PageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, "database error")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.log.Warn("Failed to close rows", zap.Error(closeErr))
		}
	}()

	var plans []*pb.TrainingPlan
	for rows.Next() {
		var plan pb.TrainingPlan
		var planDataJSON []byte
		var generatedAt, startDate, endDate time.Time

		if err := rows.Scan(&plan.Id, &plan.UserId, &planDataJSON, &generatedAt, &startDate, &endDate, &plan.Status); err != nil {
			s.log.Error("Failed to scan plan", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to read plan data")
		}

		var planData map[string]interface{}
		if err := json.Unmarshal(planDataJSON, &planData); err != nil {
			s.log.Error("Failed to unmarshal plan data", zap.Error(err))
			return nil, status.Error(codes.Internal, "invalid plan data")
		}
		planStruct, err := structpb.NewStruct(planData)
		if err != nil {
			s.log.Error("Failed to create plan struct", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to process plan data")
		}
		plan.PlanData = planStruct
		plan.GeneratedAt = timestamppb.New(generatedAt)
		plan.StartDate = timestamppb.New(startDate)
		plan.EndDate = timestamppb.New(endDate)

		plans = append(plans, &plan)
	}

	// Проверяем ошибку итерации
	if err := rows.Err(); err != nil {
		s.log.Error("Row iteration error", zap.Error(err))
		return nil, status.Error(codes.Internal, "error reading plans")
	}

	var total int32
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM training_plans WHERE user_id = $1", req.UserId).Scan(&total); err != nil {
		s.log.Warn("Failed to count plans", zap.Error(err))
		// Не блокируем ответ, просто логируем
	}

	return &pb.ListPlansResponse{
		Plans: plans,
		Total: total,
	}, nil
}

func (s *trainingServer) CompleteWorkout(ctx context.Context, req *pb.CompleteWorkoutRequest) (*pb.CompleteWorkoutResponse, error) {
	s.log.Info("CompleteWorkout",
		zap.String("user_id", req.UserId),
		zap.String("plan_id", req.PlanId),
		zap.String("workout_id", req.WorkoutId),
	)

	// Валидация входных данных
	if err := validator.ValidateCompleteWorkoutRequest(req); err != nil {
		s.log.Warn("Invalid complete workout request", zap.Error(err))
		return nil, err
	}

	// Санитизируем feedback
	feedback := sanitize.String(req.Feedback)

	// Проверяем, есть ли уже запись о выполнении
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM workout_completions
					  WHERE user_id = $1 AND training_plan_id = $2 AND workout_id = $3)
	`, req.UserId, req.PlanId, req.WorkoutId).Scan(&exists)
	if err != nil {
		s.log.Error("Failed to check existing completion", zap.Error(err))
		return nil, status.Error(codes.Internal, "database error")
	}

	if exists {
		return &pb.CompleteWorkoutResponse{Success: false}, nil
	}

	// Сохраняем выполнение (без scheduled_date, она будет NOW() по умолчанию)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO workout_completions (user_id, training_plan_id, workout_id, completed, completed_at, feedback)
		VALUES ($1, $2, $3, true, NOW(), $4)
	`, req.UserId, req.PlanId, req.WorkoutId, feedback)
	if err != nil {
		s.log.Error("Failed to save completion", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to save completion")
	}

	// Подсчитываем количество завершенных тренировок
	var completedCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM workout_completions
		WHERE user_id = $1 AND completed = true
	`, req.UserId).Scan(&completedCount)
	if err != nil {
		s.log.Error("Failed to count completions", zap.Error(err))
		completedCount = 0
	}

	var achievementID string
	switch completedCount {
	case 1:
		achievementID = "first_workout"
	case 10:
		achievementID = "ten_workouts"
	case 50:
		achievementID = "fifty_workouts"
	}

	return &pb.CompleteWorkoutResponse{
		Success:       true,
		AchievementId: achievementID,
	}, nil
}

func (s *trainingServer) GetProgress(ctx context.Context, req *pb.GetProgressRequest) (*pb.GetProgressResponse, error) {
	if err := validator.ValidateGetProgressRequest(req); err != nil {
		s.log.Warn("Invalid get progress request", zap.Error(err))
		return nil, err
	}

	s.log.Debug("GetProgress", zap.String("user_id", req.UserId))

	var totalWorkouts, completedWorkouts int32
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN completed THEN 1 END) as completed
		FROM workout_completions
		WHERE user_id = $1
	`, req.UserId).Scan(&totalWorkouts, &completedWorkouts)
	if err != nil {
		s.log.Error("Failed to get progress data", zap.Error(err))
		return nil, status.Error(codes.Internal, "database error")
	}

	completionRate := 0.0
	if totalWorkouts > 0 {
		completionRate = float64(completedWorkouts) / float64(totalWorkouts) * 100
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT workout_id, scheduled_date, completed_at
		FROM workout_completions
		WHERE user_id = $1 AND completed = true
		ORDER BY completed_at DESC
		LIMIT 20
	`, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "database error")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.log.Warn("Failed to close rows", zap.Error(closeErr))
		}
	}()

	var history []*pb.WorkoutCompletion
	for rows.Next() {
		var wc pb.WorkoutCompletion
		var scheduledDate, completedAt time.Time
		if err := rows.Scan(&wc.WorkoutId, &scheduledDate, &completedAt); err != nil {
			s.log.Error("Failed to scan workout completion", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to read workout data")
		}
		wc.ScheduledDate = timestamppb.New(scheduledDate)
		wc.CompletedAt = timestamppb.New(completedAt)
		history = append(history, &wc)
	}

	// Проверяем ошибку итерации
	if err := rows.Err(); err != nil {
		s.log.Error("Row iteration error", zap.Error(err))
		return nil, status.Error(codes.Internal, "error reading workout history")
	}

	return &pb.GetProgressResponse{
		TotalWorkouts:     totalWorkouts,
		CompletedWorkouts: completedWorkouts,
		CompletionRate:    completionRate,
		History:           history,
	}, nil
}

func main() {
	log := logger.New("training-service")
	defer log.Sync() //nolint:errcheck

	port := os.Getenv("TRAINING_SERVICE_PORT")
	if port == "" {
		port = "50053"
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
	defer func() {
		if err := database.Close(); err != nil {
			log.Error("Failed to close database", zap.Error(err))
		}
	}()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	queueName := "training_events"
	var rabbitQueue queue.Publisher
	if rabbitURL != "" {
		rabbitQueue, err = queue.NewPublisher(rabbitURL, queueName, zap.NewNop())
		if err != nil {
			log.Warn("Failed to connect to RabbitMQ", zap.Error(err))
		} else {
			defer func() { _ = rabbitQueue.Close() }()
			log.Info("RabbitMQ connected", zap.String("queue", queueName))
		}
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	pb.RegisterTrainingServiceServer(s, &trainingServer{
		db:          database,
		log:         log,
		rabbitQueue: rabbitQueue,
	})

	log.Info("Training service starting", zap.String("port", port))
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve", zap.Error(err))
	}
}
