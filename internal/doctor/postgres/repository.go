// Package postgres implements the doctor service repository ports using PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MAMUER/Project/internal/doctor/model"
	"github.com/MAMUER/Project/internal/doctor/port"
)

// Ensure compile-time interface compliance
var (
	_ port.DoctorRepository               = (*DoctorRepo)(nil)
	_ port.SubscriptionRepository         = (*SubscriptionRepo)(nil)
	_ port.MessageRepository              = (*MessageRepo)(nil)
	_ port.PrescriptionRepository         = (*PrescriptionRepo)(nil)
	_ port.ConsultationRepository         = (*ConsultationRepo)(nil)
	_ port.TrainingModificationRepository = (*TrainingModificationRepo)(nil)
)

// DoctorRepo implements port.DoctorRepository
type DoctorRepo struct {
	db *sql.DB
}

func NewDoctorRepo(db *sql.DB) *DoctorRepo {
	return &DoctorRepo{db: db}
}

func (r *DoctorRepo) List(ctx context.Context, specialty string, page, pageSize int32) ([]model.Doctor, int64, error) {
	args := []interface{}{}
	whereClause := "WHERE d.is_active = true"
	argIdx := 1

	if specialty != "" {
		whereClause += fmt.Sprintf(" AND d.specialty = $%d", argIdx)
		args = append(args, specialty)
		argIdx++
	}

	offset := (page - 1) * pageSize

	var query strings.Builder
	query.WriteString(`
		SELECT d.id, COALESCE(d.user_id::text, ''), COALESCE(d.specialty, ''),
		       COALESCE(d.license_number, ''), COALESCE(d.phone, ''), COALESCE(d.bio, ''),
		       d.is_active, d.created_at, d.updated_at
		FROM doctors d
		`)
	query.WriteString(whereClause)
	fmt.Fprintf(&query, `
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d
	`, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query doctors: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var doctors []model.Doctor
	for rows.Next() {
		var d model.Doctor
		var userID sql.NullString
		if err := rows.Scan(&d.ID, &userID, &d.Specialty, &d.LicenseNumber, &d.Phone, &d.Bio, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan doctor: %w", err)
		}
		if userID.Valid {
			s := userID.String
			d.UserID = &s
		}
		doctors = append(doctors, d)
	}

	// Count total
	var total int64
	countQuery := "SELECT COUNT(*) FROM doctors WHERE is_active = true"
	if specialty != "" {
		countQuery += " AND specialty = $1"
		if err := r.db.QueryRowContext(ctx, countQuery, specialty).Scan(&total); err != nil {
			total = int64(len(doctors))
		}
	} else {
		if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM doctors WHERE is_active = true").Scan(&total); err != nil {
			total = int64(len(doctors))
		}
	}

	return doctors, total, nil
}

func (r *DoctorRepo) GetByID(ctx context.Context, id string) (*model.Doctor, error) {
	var d model.Doctor
	var userID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(user_id::text, ''), COALESCE(specialty, ''),
		       COALESCE(license_number, ''), COALESCE(phone, ''), COALESCE(bio, ''),
		       is_active, created_at, updated_at
		FROM doctors WHERE id = $1
	`, id).Scan(&d.ID, &userID, &d.Specialty, &d.LicenseNumber, &d.Phone, &d.Bio, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("doctor not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get doctor: %w", err)
	}
	if userID.Valid {
		s := userID.String
		d.UserID = &s
	}
	return &d, nil
}

func (r *DoctorRepo) GetByUserID(ctx context.Context, userID string) (*model.Doctor, error) {
	var d model.Doctor
	err := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(user_id::text, ''), COALESCE(specialty, ''),
		       COALESCE(license_number, ''), COALESCE(phone, ''), COALESCE(bio, ''),
		       is_active, created_at, updated_at
		FROM doctors WHERE user_id = $1
	`, userID).Scan(&d.ID, &d.UserID, &d.Specialty, &d.LicenseNumber, &d.Phone, &d.Bio, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("doctor not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get doctor by user_id: %w", err)
	}
	return &d, nil
}

func (r *DoctorRepo) Create(ctx context.Context, d *model.Doctor) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO doctors (user_id, specialty, license_number, phone, bio, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at
	`, d.UserID, d.Specialty, d.LicenseNumber, d.Phone, d.Bio, d.IsActive).Scan(&d.ID, &d.CreatedAt)
}

func (r *DoctorRepo) Update(ctx context.Context, d *model.Doctor) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE doctors SET specialty = $2, license_number = $3, phone = $4, bio = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1
	`, d.ID, d.Specialty, d.LicenseNumber, d.Phone, d.Bio, d.IsActive)
	return err
}

// SubscriptionRepo implements port.SubscriptionRepository
type SubscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *model.Subscription) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO doctor_subscriptions (id, user_id, doctor_id, plan_type, starts_at, expires_at, is_active, price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`, s.ID, s.UserID, s.DoctorID, s.PlanType, s.StartsAt, s.ExpiresAt, s.IsActive, s.Price, s.CreatedAt).Scan(&s.CreatedAt)
}

func (r *SubscriptionRepo) GetByUserAndDoctor(ctx context.Context, userID, doctorID string) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, doctor_id, plan_type, starts_at, expires_at, is_active, price, created_at
		FROM doctor_subscriptions
		WHERE user_id = $1 AND doctor_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, userID, doctorID).Scan(&s.ID, &s.UserID, &s.DoctorID, &s.PlanType, &s.StartsAt, &s.ExpiresAt, &s.IsActive, &s.Price, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription not found")
	}
	return &s, err
}

func (r *SubscriptionRepo) IsActive(ctx context.Context, userID, doctorID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM doctor_subscriptions
			WHERE user_id = $1 AND doctor_id = $2 AND is_active = true AND expires_at > NOW()
		)
	`, userID, doctorID).Scan(&exists)
	return exists, err
}

func (r *SubscriptionRepo) Cancel(ctx context.Context, userID, doctorID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE doctor_subscriptions SET is_active = false
		WHERE user_id = $1 AND doctor_id = $2 AND is_active = true
	`, userID, doctorID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("active subscription not found")
	}
	return nil
}

// MessageRepo implements port.MessageRepository
type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(ctx context.Context, m *model.Message) (*model.Message, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO consultation_messages (id, user_id, doctor_id, sender_user_id, sender_doctor_id, message, message_type, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`, m.ID, m.UserID, m.DoctorID, m.SenderUserID, m.SenderDoctorID, m.Message, m.MessageType, m.IsRead, m.CreatedAt).Scan(&m.CreatedAt)
	return m, err
}

func (r *MessageRepo) GetChatHistory(ctx context.Context, userID, doctorID string, page, pageSize int32) ([]model.Message, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, doctor_id, sender_user_id, sender_doctor_id, message, message_type, is_read, created_at
		FROM consultation_messages
		WHERE user_id = $1 AND doctor_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, doctorID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.DoctorID, &m.SenderUserID, &m.SenderDoctorID, &m.Message, &m.MessageType, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (r *MessageRepo) MarkAsRead(ctx context.Context, userID, doctorID string) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE consultation_messages SET is_read = true
		WHERE user_id = $1 AND doctor_id = $2 AND is_read = false
	`, userID, doctorID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *MessageRepo) GetUnreadCount(ctx context.Context, userID string) (map[string]int64, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT doctor_id, COUNT(*) as cnt
		FROM consultation_messages
		WHERE user_id = $1 AND is_read = false
		GROUP BY doctor_id
	`, userID)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	unreadByDoctor := make(map[string]int64)
	var total int64
	for rows.Next() {
		var doctorID string
		var count int64
		if err := rows.Scan(&doctorID, &count); err != nil {
			continue
		}
		unreadByDoctor[doctorID] = count
		total += count
	}
	return unreadByDoctor, total, nil
}

// PrescriptionRepo implements port.PrescriptionRepository
type PrescriptionRepo struct {
	db *sql.DB
}

func NewPrescriptionRepo(db *sql.DB) *PrescriptionRepo {
	return &PrescriptionRepo{db: db}
}

func (r *PrescriptionRepo) Create(ctx context.Context, p *model.Prescription) (*model.Prescription, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO doctor_prescriptions (id, consultation_id, user_id, doctor_id, prescription_type, title, description, priority, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`, p.ID, p.ConsultationID, p.UserID, p.DoctorID, p.PrescriptionType, p.Title, p.Description, p.Priority, p.Status, p.CreatedAt, p.UpdatedAt).Scan(&p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *PrescriptionRepo) GetByUser(ctx context.Context, userID, statusFilter string) ([]model.Prescription, error) {
	query := `
		SELECT id, COALESCE(consultation_id::text, ''), user_id, doctor_id, prescription_type, title, COALESCE(description, ''), priority, status, created_at, updated_at
		FROM doctor_prescriptions WHERE user_id = $1
	`
	args := []interface{}{userID}
	if statusFilter != "" && statusFilter != "all" {
		query += " AND status = $2"
		args = append(args, statusFilter)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var prescriptions []model.Prescription
	for rows.Next() {
		var p model.Prescription
		var consultID sql.NullString
		if err := rows.Scan(&p.ID, &consultID, &p.UserID, &p.DoctorID, &p.PrescriptionType, &p.Title, &p.Description, &p.Priority, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if consultID.Valid {
			s := consultID.String
			p.ConsultationID = &s
		}
		prescriptions = append(prescriptions, p)
	}
	return prescriptions, nil
}

func (r *PrescriptionRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE doctor_prescriptions SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, status)
	return err
}

// ConsultationRepo implements port.ConsultationRepository
type ConsultationRepo struct {
	db *sql.DB
}

func NewConsultationRepo(db *sql.DB) *ConsultationRepo {
	return &ConsultationRepo{db: db}
}

func (r *ConsultationRepo) Create(ctx context.Context, c *model.Consultation) (*model.Consultation, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO consultations (id, user_id, doctor_id, status, scheduled_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`, c.ID, c.UserID, c.DoctorID, c.Status, c.ScheduledAt, c.CreatedAt).Scan(&c.CreatedAt)
	return c, err
}

func (r *ConsultationRepo) GetByUser(ctx context.Context, userID, statusFilter string) ([]model.Consultation, error) {
	query := `
		SELECT id, user_id, doctor_id, status, scheduled_at, COALESCE(started_at, '1970-01-01'::timestamptz), COALESCE(ended_at, '1970-01-01'::timestamptz), COALESCE(notes, ''), created_at
		FROM consultations WHERE user_id = $1
	`
	args := []interface{}{userID}
	if statusFilter != "" && statusFilter != "all" {
		query += " AND status = $2"
		args = append(args, statusFilter)
	}
	query += " ORDER BY scheduled_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var consultations []model.Consultation
	for rows.Next() {
		var c model.Consultation
		var notes string
		if err := rows.Scan(&c.ID, &c.UserID, &c.DoctorID, &c.Status, &c.ScheduledAt, &c.StartedAt, &c.EndedAt, &notes, &c.CreatedAt); err != nil {
			return nil, err
		}
		if notes != "" {
			c.Notes = notes
		}
		consultations = append(consultations, c)
	}
	return consultations, nil
}

func (r *ConsultationRepo) Complete(ctx context.Context, id, notes string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE consultations SET status = 'completed', ended_at = NOW(), notes = $2
		WHERE id = $1 AND status IN ('scheduled', 'in_progress')
	`, id, notes)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("consultation not found or already completed")
	}
	return nil
}

// TrainingModificationRepo implements port.TrainingModificationRepository
type TrainingModificationRepo struct {
	db *sql.DB
}

func NewTrainingModificationRepo(db *sql.DB) *TrainingModificationRepo {
	return &TrainingModificationRepo{db: db}
}

func (r *TrainingModificationRepo) Create(ctx context.Context, tm *model.TrainingModification) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO doctor_training_modifications (id, doctor_id, training_plan_id, modification_type, old_value, new_value, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`, tm.ID, tm.DoctorID, tm.TrainingPlanID, tm.ModificationType, tm.OldValue, tm.NewValue, tm.Reason, tm.CreatedAt).Scan(&tm.CreatedAt)
}

func (r *TrainingModificationRepo) GetByUser(ctx context.Context, userID, trainingPlanID string) ([]model.TrainingModification, error) {
	// Note: user_id is available via JOIN with training_plans
	query := `
		SELECT tm.id, tm.doctor_id, tm.training_plan_id, tm.modification_type,
		       COALESCE(tm.old_value::text, ''), COALESCE(tm.new_value::text, ''),
		       COALESCE(tm.reason, ''), tm.created_at
		FROM doctor_training_modifications tm
		JOIN training_plans tp ON tp.id = tm.training_plan_id
		WHERE tp.user_id = $1
	`
	args := []interface{}{userID}
	if trainingPlanID != "" {
		query += " AND tm.training_plan_id = $2"
		args = append(args, trainingPlanID)
	}
	query += " ORDER BY tm.created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var modifications []model.TrainingModification
	for rows.Next() {
		var tm model.TrainingModification
		if err := rows.Scan(&tm.ID, &tm.DoctorID, &tm.TrainingPlanID, &tm.ModificationType, &tm.OldValue, &tm.NewValue, &tm.Reason, &tm.CreatedAt); err != nil {
			return nil, err
		}
		modifications = append(modifications, tm)
	}
	return modifications, nil
}
