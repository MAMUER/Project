// cmd/doctor-service/main.go — Doctor Consultant Service
//
// Сервис для консульттаций врача:
// • Подписки на врача
// • Чат между врачом и пациентом
// • Назначения врача (рецепты, рекомендации)
// • Редактирование тренировок врачом
// • История консультаций

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MAMUER/Project/api/gen/doctor"
	"github.com/MAMUER/Project/internal/db"
	"github.com/MAMUER/Project/internal/logger"
)

type doctorService struct {
	doctor.UnimplementedDoctorServiceServer
	db  *sql.DB
	log *logger.Logger
}

func main() {
	log := logger.New("doctor-service")
	defer func() {
		if syncErr := log.Sync(); syncErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", syncErr)
		}
	}()

	port := os.Getenv("DOCTOR_SERVICE_PORT")
	if port == "" {
		port = "50054"
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
		if closeErr := database.Close(); closeErr != nil {
			log.Error("Failed to close database connection", zap.Error(closeErr))
		}
	}()

	s := &doctorService{
		db:  database,
		log: log,
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	doctor.RegisterDoctorServiceServer(grpcServer, s)

	log.Info("Doctor service starting", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve", zap.Error(err))
	}
}

// ===== Doctors =====

func (s *doctorService) ListDoctors(ctx context.Context, req *doctor.ListDoctorsRequest) (*doctor.ListDoctorsResponse, error) {
	query := `
		SELECT id, email, full_name, specialty, license_number, phone, bio, rating, consultation_count, is_active
		FROM doctors
		WHERE is_active = true
	`

	args := []interface{}{}
	if req.GetSpecialty() != "" {
		query += " AND specialty = $1"
		args = append(args, req.GetSpecialty())
	}

	query += " ORDER BY rating DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, req.GetPageSize(), (req.GetPage()-1)*req.GetPageSize())

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.log.Error("Failed to query doctors", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	defer rows.Close()

	var doctors []*doctor.Doctor
	for rows.Next() {
		var d doctor.Doctor
		if err := rows.Scan(
			&d.Id, &d.Email, &d.FullName, &d.Specialty, &d.LicenseNumber,
			&d.Phone, &d.Bio, &d.Rating, &d.ConsultationCount, &d.IsActive,
		); err != nil {
			s.log.Error("Failed to scan doctor", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}
		doctors = append(doctors, &d)
	}

	// Count total
	var total int32
	countQuery := "SELECT COUNT(*) FROM doctors WHERE is_active = true"
	if req.GetSpecialty() != "" {
		countQuery += " AND specialty = $1"
		if err := s.db.QueryRowContext(ctx, countQuery, req.GetSpecialty()).Scan(&total); err != nil {
			total = int32(len(doctors))
		}
	} else {
		if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM doctors WHERE is_active = true").Scan(&total); err != nil {
			total = int32(len(doctors))
		}
	}

	return &doctor.ListDoctorsResponse{
		Doctors: doctors,
		Total:   total,
	}, nil
}

func (s *doctorService) GetDoctor(ctx context.Context, req *doctor.GetDoctorRequest) (*doctor.Doctor, error) {
	var d doctor.Doctor
	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, full_name, specialty, license_number, phone, bio, rating, consultation_count, is_active
		FROM doctors
		WHERE id = $1
	`, req.GetDoctorId()).Scan(
		&d.Id, &d.Email, &d.FullName, &d.Specialty, &d.LicenseNumber,
		&d.Phone, &d.Bio, &d.Rating, &d.ConsultationCount, &d.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "doctor not found")
	}
	if err != nil {
		s.log.Error("Failed to get doctor", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &d, nil
}

// ===== Subscriptions =====

func (s *doctorService) SubscribeToDoctor(ctx context.Context, req *doctor.SubscribeRequest) (*doctor.SubscribeResponse, error) {
	// Проверяем, существует ли врач
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM doctors WHERE id = $1 AND is_active = true)", req.GetDoctorId()).Scan(&exists); err != nil || !exists {
		return nil, status.Error(codes.NotFound, "doctor not found or inactive")
	}

	// Определяем срок действия
	var duration time.Duration
	switch req.GetPlanType() {
	case "monthly":
		duration = 30 * 24 * time.Hour
	case "quarterly":
		duration = 90 * 24 * time.Hour
	case "yearly":
		duration = 365 * 24 * time.Hour
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid plan type")
	}

	// Цены
	prices := map[string]float64{
		"monthly":   999.0,
		"quarterly": 2499.0,
		"yearly":    7999.0,
	}
	price := prices[req.GetPlanType()]

	subscriptionID := uuid.New().String()
	expiresAt := time.Now().Add(duration)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO doctor_subscriptions (id, user_id, doctor_id, plan_type, starts_at, expires_at, price, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		ON CONFLICT (user_id, doctor_id) DO UPDATE
		SET plan_type = $4, starts_at = $5, expires_at = $6, price = $7, is_active = true
	`, subscriptionID, req.GetUserId(), req.GetDoctorId(), req.GetPlanType(), time.Now(), expiresAt, price)

	if err != nil {
		s.log.Error("Failed to create subscription", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create subscription")
	}

	return &doctor.SubscribeResponse{
		SubscriptionId: subscriptionID,
		ExpiresAt:      timestamppb.New(expiresAt),
		Price:          price,
	}, nil
}

func (s *doctorService) GetSubscription(ctx context.Context, req *doctor.GetSubscriptionRequest) (*doctor.Subscription, error) {
	var sub doctor.Subscription
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, doctor_id, plan_type, starts_at, expires_at, is_active, price
		FROM doctor_subscriptions
		WHERE user_id = $1 AND doctor_id = $2
	`, req.GetUserId(), req.GetDoctorId()).Scan(
		&sub.Id, &sub.UserId, &sub.DoctorId, &sub.PlanType,
		&sub.StartsAt, &sub.ExpiresAt, &sub.IsActive, &sub.Price,
	)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "subscription not found")
	}
	if err != nil {
		s.log.Error("Failed to get subscription", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &sub, nil
}

func (s *doctorService) CancelSubscription(ctx context.Context, req *doctor.CancelSubscriptionRequest) (*doctor.CancelSubscriptionResponse, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE doctor_subscriptions
		SET is_active = false
		WHERE user_id = $1 AND doctor_id = $2 AND is_active = true
	`, req.GetUserId(), req.GetDoctorId())

	if err != nil {
		s.log.Error("Failed to cancel subscription", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, status.Error(codes.NotFound, "active subscription not found")
	}

	return &doctor.CancelSubscriptionResponse{
		Success: true,
		Message: "Подписка отменена",
	}, nil
}

// ===== Messages =====

func (s *doctorService) SendMessage(ctx context.Context, req *doctor.SendMessageRequest) (*doctor.Message, error) {
	// Проверяем подписку
	var hasSubscription bool
	if err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM doctor_subscriptions
			WHERE user_id = $1 AND doctor_id = $2 AND is_active = true AND expires_at > NOW()
		)
	`, req.GetUserId(), req.GetDoctorId()).Scan(&hasSubscription); err != nil || !hasSubscription {
		return nil, status.Error(codes.PermissionDenied, "active subscription required")
	}

	messageID := uuid.New().String()
	var msg doctor.Message
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO consultation_messages (id, user_id, doctor_id, sender_id, sender_type, message, message_type, is_read)
		VALUES ($1, $2, $3, $4, $5, $6, $7, false)
		RETURNING id, user_id, doctor_id, sender_id, sender_type, message, message_type, is_read, created_at
	`, messageID, req.GetUserId(), req.GetDoctorId(), req.GetSenderId(), req.GetSenderType(), req.GetMessage(), req.GetMessageType()).Scan(
		&msg.Id, &msg.UserId, &msg.DoctorId, &msg.SenderId, &msg.SenderType,
		&msg.Message, &msg.MessageType, &msg.IsRead, &msg.CreatedAt,
	)

	if err != nil {
		s.log.Error("Failed to send message", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &msg, nil
}

func (s *doctorService) GetChatHistory(ctx context.Context, req *doctor.GetChatHistoryRequest) (*doctor.GetChatHistoryResponse, error) {
	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 50
	}
	offset := (req.GetPage() - 1) * pageSize

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, doctor_id, sender_id, sender_type, message, message_type, is_read, created_at
		FROM consultation_messages
		WHERE user_id = $1 AND doctor_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, req.GetUserId(), req.GetDoctorId(), pageSize, offset)

	if err != nil {
		s.log.Error("Failed to get chat history", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	defer rows.Close()

	var messages []*doctor.Message
	for rows.Next() {
		var msg doctor.Message
		if err := rows.Scan(
			&msg.Id, &msg.UserId, &msg.DoctorId, &msg.SenderId, &msg.SenderType,
			&msg.Message, &msg.MessageType, &msg.IsRead, &msg.CreatedAt,
		); err != nil {
			s.log.Error("Failed to scan message", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}
		messages = append(messages, &msg)
	}

	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return &doctor.GetChatHistoryResponse{
		Messages: messages,
		Total:    int32(len(messages)),
	}, nil
}

func (s *doctorService) MarkMessagesRead(ctx context.Context, req *doctor.MarkMessagesReadRequest) (*doctor.MarkMessagesReadResponse, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE consultation_messages
		SET is_read = true
		WHERE user_id = $1 AND doctor_id = $2 AND is_read = false
	`, req.GetUserId(), req.GetDoctorId())

	if err != nil {
		s.log.Error("Failed to mark messages read", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	rows, _ := result.RowsAffected()
	return &doctor.MarkMessagesReadResponse{
		MarkedCount: int32(rows),
	}, nil
}

func (s *doctorService) GetUnreadCount(ctx context.Context, req *doctor.GetUnreadCountRequest) (*doctor.GetUnreadCountResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT doctor_id, COUNT(*) as cnt
		FROM consultation_messages
		WHERE user_id = $1 AND is_read = false
		GROUP BY doctor_id
	`, req.GetUserId())

	if err != nil {
		s.log.Error("Failed to get unread count", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	defer rows.Close()

	unreadByDoctor := make(map[string]int32)
	var total int32
	for rows.Next() {
		var doctorID string
		var count int32
		if err := rows.Scan(&doctorID, &count); err != nil {
			s.log.Error("Failed to scan unread count", zap.Error(err))
			continue
		}
		unreadByDoctor[doctorID] = count
		total += count
	}

	return &doctor.GetUnreadCountResponse{
		UnreadByDoctor: unreadByDoctor,
		TotalUnread:    total,
	}, nil
}

// ===== Prescriptions =====

func (s *doctorService) CreatePrescription(ctx context.Context, req *doctor.CreatePrescriptionRequest) (*doctor.Prescription, error) {
	prescriptionID := uuid.New().String()
	var presc doctor.Prescription
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO doctor_prescriptions (id, user_id, doctor_id, prescription_type, title, description, priority, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		RETURNING id, user_id, doctor_id, prescription_type, title, description, priority, status, created_at, updated_at
	`, prescriptionID, req.GetUserId(), req.GetDoctorId(), req.GetPrescriptionType(), req.GetTitle(), req.GetDescription(), req.GetPriority()).Scan(
		&presc.Id, &presc.UserId, &presc.DoctorId, &presc.PrescriptionType,
		&presc.Title, &presc.Description, &presc.Priority, &presc.Status,
		&presc.CreatedAt, &presc.UpdatedAt,
	)

	if err != nil {
		s.log.Error("Failed to create prescription", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &presc, nil
}

func (s *doctorService) GetPrescriptions(ctx context.Context, req *doctor.GetPrescriptionsRequest) (*doctor.GetPrescriptionsResponse, error) {
	query := `
		SELECT id, user_id, doctor_id, prescription_type, title, description, priority, status, created_at, updated_at
		FROM doctor_prescriptions
		WHERE user_id = $1
	`
	args := []interface{}{req.GetUserId()}

	if req.GetStatusFilter() != "" && req.GetStatusFilter() != "all" {
		query += " AND status = $2"
		args = append(args, req.GetStatusFilter())
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.log.Error("Failed to get prescriptions", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	defer rows.Close()

	var prescriptions []*doctor.Prescription
	for rows.Next() {
		var p doctor.Prescription
		if err := rows.Scan(
			&p.Id, &p.UserId, &p.DoctorId, &p.PrescriptionType, &p.Title,
			&p.Description, &p.Priority, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			s.log.Error("Failed to scan prescription", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}
		prescriptions = append(prescriptions, &p)
	}

	return &doctor.GetPrescriptionsResponse{
		Prescriptions: prescriptions,
		Total:         int32(len(prescriptions)),
	}, nil
}

func (s *doctorService) UpdatePrescriptionStatus(ctx context.Context, req *doctor.UpdatePrescriptionStatusRequest) (*doctor.UpdatePrescriptionStatusResponse, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE doctor_prescriptions
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, req.GetPrescriptionId(), req.GetNewStatus())

	if err != nil {
		s.log.Error("Failed to update prescription status", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, status.Error(codes.NotFound, "prescription not found")
	}

	return &doctor.UpdatePrescriptionStatusResponse{
		Success: true,
		Message: "Статус обновлен",
	}, nil
}

// ===== Training Modifications =====

func (s *doctorService) ModifyTrainingPlan(ctx context.Context, req *doctor.ModifyTrainingPlanRequest) (*doctor.ModifyTrainingPlanResponse, error) {
	modificationID := uuid.New().String()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO doctor_training_modifications (id, user_id, doctor_id, training_plan_id, modification_type, old_value, new_value, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, modificationID, req.GetUserId(), req.GetDoctorId(), req.GetTrainingPlanId(), req.GetModificationType(), req.GetOldValue(), req.GetNewValue(), req.GetReason())

	if err != nil {
		s.log.Error("Failed to create training modification", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	// TODO: Here we would also update the actual training plan in training_plans table
	// For now we just log the modification

	return &doctor.ModifyTrainingPlanResponse{
		ModificationId: modificationID,
		Success:        true,
		Message:        "Тренировочный план обновлен",
	}, nil
}

func (s *doctorService) GetTrainingModifications(ctx context.Context, req *doctor.GetTrainingModificationsRequest) (*doctor.GetTrainingModificationsResponse, error) {
	query := `
		SELECT id, user_id, doctor_id, training_plan_id, modification_type, old_value, new_value, reason, created_at
		FROM doctor_training_modifications
		WHERE user_id = $1
	`
	args := []interface{}{req.GetUserId()}

	if req.GetTrainingPlanId() != "" {
		query += " AND training_plan_id = $2"
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.log.Error("Failed to get modifications", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	defer rows.Close()

	var modifications []*doctor.TrainingModification
	for rows.Next() {
		var m doctor.TrainingModification
		if err := rows.Scan(
			&m.Id, &m.UserId, &m.DoctorId, &m.TrainingPlanId, &m.ModificationType,
			&m.OldValue, &m.NewValue, &m.Reason, &m.CreatedAt,
		); err != nil {
			s.log.Error("Failed to scan modification", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}
		modifications = append(modifications, &m)
	}

	return &doctor.GetTrainingModificationsResponse{
		Modifications: modifications,
		Total:         int32(len(modifications)),
	}, nil
}

// ===== Consultations =====

func (s *doctorService) ScheduleConsultation(ctx context.Context, req *doctor.ScheduleConsultationRequest) (*doctor.Consultation, error) {
	consultationID := uuid.New().String()

	// Проверяем подписку
	var hasSubscription bool
	if err := s.db.QueryRowContext(ctx, `
		SELECT has_active_subscription($1, $2)
	`, req.GetUserId(), req.GetDoctorId()).Scan(&hasSubscription); err != nil || !hasSubscription {
		return nil, status.Error(codes.PermissionDenied, "active subscription required")
	}

	var consult doctor.Consultation
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO consultations (id, user_id, doctor_id, status, scheduled_at)
		VALUES ($1, $2, $3, 'scheduled', $4)
		RETURNING id, user_id, doctor_id, status, scheduled_at, created_at
	`, consultationID, req.GetUserId(), req.GetDoctorId(), req.GetScheduledAt().AsTime()).Scan(
		&consult.Id, &consult.UserId, &consult.DoctorId, &consult.Status,
		&consult.ScheduledAt, &consult.CreatedAt,
	)

	if err != nil {
		s.log.Error("Failed to schedule consultation", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &consult, nil
}

func (s *doctorService) GetConsultations(ctx context.Context, req *doctor.GetConsultationsRequest) (*doctor.GetConsultationsResponse, error) {
	query := `
		SELECT id, user_id, doctor_id, status, scheduled_at, started_at, ended_at, notes, created_at
		FROM consultations
		WHERE user_id = $1
	`
	args := []interface{}{req.GetUserId()}

	if req.GetStatusFilter() != "" && req.GetStatusFilter() != "all" {
		query += " AND status = $2"
		args = append(args, req.GetStatusFilter())
	}

	query += " ORDER BY scheduled_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.log.Error("Failed to get consultations", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	defer rows.Close()

	var consultations []*doctor.Consultation
	for rows.Next() {
		var c doctor.Consultation
		var notes sql.NullString
		if err := rows.Scan(
			&c.Id, &c.UserId, &c.DoctorId, &c.Status, &c.ScheduledAt,
			&c.StartedAt, &c.EndedAt, &notes, &c.CreatedAt,
		); err != nil {
			s.log.Error("Failed to scan consultation", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}
		if notes.Valid {
			c.Notes = notes.String
		}
		consultations = append(consultations, &c)
	}

	return &doctor.GetConsultationsResponse{
		Consultations: consultations,
		Total:         int32(len(consultations)),
	}, nil
}

func (s *doctorService) CompleteConsultation(ctx context.Context, req *doctor.CompleteConsultationRequest) (*doctor.CompleteConsultationResponse, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE consultations
		SET status = 'completed', ended_at = NOW(), notes = $2
		WHERE id = $1 AND status IN ('scheduled', 'in_progress')
	`, req.GetConsultationId(), req.GetNotes())

	if err != nil {
		s.log.Error("Failed to complete consultation", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, status.Error(codes.NotFound, "consultation not found or already completed")
	}

	return &doctor.CompleteConsultationResponse{
		Success: true,
		Message: "Консультация завершена",
	}, nil
}
