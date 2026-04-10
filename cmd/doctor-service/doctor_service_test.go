// cmd/doctor-service/doctor_service_test.go
package main

import (
	"context"
	"testing"
	"time"

	"github.com/MAMUER/Project/internal/doctor/model"
	"github.com/MAMUER/Project/internal/doctor/service"
)

// ===== In-memory mock repositories =====

type mockDoctorRepo struct {
	doctors []model.Doctor
}

func (m *mockDoctorRepo) List(ctx context.Context, specialty string, page, pageSize int32) ([]model.Doctor, int64, error) {
	return m.doctors, int64(len(m.doctors)), nil
}
func (m *mockDoctorRepo) GetByID(ctx context.Context, id string) (*model.Doctor, error) {
	for _, d := range m.doctors {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, nil
}
func (m *mockDoctorRepo) GetByUserID(ctx context.Context, userID string) (*model.Doctor, error) {
	return nil, nil
}
func (m *mockDoctorRepo) Create(ctx context.Context, d *model.Doctor) error { return nil }
func (m *mockDoctorRepo) Update(ctx context.Context, d *model.Doctor) error { return nil }

type mockSubscriptionRepo struct {
	subs        []*model.Subscription
	activeCheck bool
	activeErr   error
}

func (m *mockSubscriptionRepo) Create(ctx context.Context, s *model.Subscription) error {
	m.subs = append(m.subs, s)
	return nil
}
func (m *mockSubscriptionRepo) GetByUserAndDoctor(ctx context.Context, userID, doctorID string) (*model.Subscription, error) {
	for _, s := range m.subs {
		if s.UserID == userID && s.DoctorID == doctorID {
			return s, nil
		}
	}
	return nil, nil
}
func (m *mockSubscriptionRepo) Cancel(ctx context.Context, userID, doctorID string) error { return nil }
func (m *mockSubscriptionRepo) IsActive(ctx context.Context, userID, doctorID string) (bool, error) {
	return m.activeCheck, m.activeErr
}

type mockMessageRepo struct {
	messages []model.Message
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *model.Message) (*model.Message, error) {
	m.messages = append(m.messages, *msg)
	return msg, nil
}
func (m *mockMessageRepo) GetChatHistory(ctx context.Context, userID, doctorID string, page, pageSize int32) ([]model.Message, error) {
	return m.messages, nil
}
func (m *mockMessageRepo) MarkAsRead(ctx context.Context, userID, doctorID string) (int64, error) {
	return 0, nil
}
func (m *mockMessageRepo) GetUnreadCount(ctx context.Context, userID string) (map[string]int64, int64, error) {
	return nil, 0, nil
}

type mockPrescriptionRepo struct{}

func (m *mockPrescriptionRepo) Create(ctx context.Context, p *model.Prescription) (*model.Prescription, error) {
	return p, nil
}
func (m *mockPrescriptionRepo) GetByUser(ctx context.Context, userID, statusFilter string) ([]model.Prescription, error) {
	return nil, nil
}
func (m *mockPrescriptionRepo) UpdateStatus(ctx context.Context, id, status string) error { return nil }

type mockConsultationRepo struct{}

func (m *mockConsultationRepo) Create(ctx context.Context, c *model.Consultation) (*model.Consultation, error) {
	return c, nil
}
func (m *mockConsultationRepo) GetByUser(ctx context.Context, userID, statusFilter string) ([]model.Consultation, error) {
	return nil, nil
}
func (m *mockConsultationRepo) Complete(ctx context.Context, id, notes string) error { return nil }

type mockModificationRepo struct{}

func (m *mockModificationRepo) Create(ctx context.Context, tm *model.TrainingModification) error {
	return nil
}
func (m *mockModificationRepo) GetByUser(ctx context.Context, userID, trainingPlanID string) ([]model.TrainingModification, error) {
	return nil, nil
}

// ===== Tests =====

func TestService_Subscribe_InvalidPlanType(t *testing.T) {
	svc := service.New(
		&mockDoctorRepo{},
		&mockSubscriptionRepo{},
		&mockMessageRepo{},
		&mockPrescriptionRepo{},
		&mockConsultationRepo{},
		&mockModificationRepo{},
	)

	_, err := svc.Subscribe(context.Background(), "user-1", "doc-1", "invalid_plan")
	if err == nil {
		t.Fatal("expected error for invalid plan type")
	}
	if err.Error() != "invalid plan type" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Subscribe_ValidPlan(t *testing.T) {
	subRepo := &mockSubscriptionRepo{}
	svc := service.New(
		&mockDoctorRepo{},
		subRepo,
		&mockMessageRepo{},
		&mockPrescriptionRepo{},
		&mockConsultationRepo{},
		&mockModificationRepo{},
	)

	sub, err := svc.Subscribe(context.Background(), "user-1", "doc-1", "monthly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.UserID != "user-1" {
		t.Errorf("expected user_id 'user-1', got '%s'", sub.UserID)
	}
	if sub.PlanType != "monthly" {
		t.Errorf("expected plan_type 'monthly', got '%s'", sub.PlanType)
	}
	if !sub.IsActive {
		t.Error("expected subscription to be active")
	}
	if sub.Price != 999.0 {
		t.Errorf("expected price 999.0, got %f", sub.Price)
	}
	// Monthly should be ~30 days
	expectedExpiry := sub.StartsAt.Add(30 * 24 * time.Hour)
	if sub.ExpiresAt.Sub(expectedExpiry) > time.Second {
		t.Errorf("expected expiry around %v, got %v", expectedExpiry, sub.ExpiresAt)
	}
}

func TestService_SendMessage_NoSubscription(t *testing.T) {
	subRepo := &mockSubscriptionRepo{activeCheck: false}
	svc := service.New(
		&mockDoctorRepo{},
		subRepo,
		&mockMessageRepo{},
		&mockPrescriptionRepo{},
		&mockConsultationRepo{},
		&mockModificationRepo{},
	)

	msg := &model.Message{
		UserID:   "user-1",
		DoctorID: "doc-1",
		Message:  "Hello",
	}
	_, err := svc.SendMessage(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error when no active subscription")
	}
	if err.Error() != "active subscription required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_SendMessage_WithSubscription(t *testing.T) {
	msgRepo := &mockMessageRepo{}
	subRepo := &mockSubscriptionRepo{activeCheck: true}
	svc := service.New(
		&mockDoctorRepo{},
		subRepo,
		msgRepo,
		&mockPrescriptionRepo{},
		&mockConsultationRepo{},
		&mockModificationRepo{},
	)

	msg := &model.Message{
		UserID:      "user-1",
		DoctorID:    "doc-1",
		Message:     "Hello doctor",
		MessageType: "text",
	}
	result, err := svc.SendMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Message != "Hello doctor" {
		t.Errorf("expected message 'Hello doctor', got '%s'", result.Message)
	}
	if result.IsRead {
		t.Error("expected new message to be unread")
	}
	if len(msgRepo.messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgRepo.messages))
	}
}

func TestService_ListDoctors(t *testing.T) {
	doctorRepo := &mockDoctorRepo{
		doctors: []model.Doctor{
			{ID: "doc-1", Specialty: "cardiology"},
			{ID: "doc-2", Specialty: "neurology"},
		},
	}
	svc := service.New(
		doctorRepo,
		&mockSubscriptionRepo{},
		&mockMessageRepo{},
		&mockPrescriptionRepo{},
		&mockConsultationRepo{},
		&mockModificationRepo{},
	)

	doctors, total, err := svc.ListDoctors(context.Background(), "", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 doctors, got %d", total)
	}
	if len(doctors) != 2 {
		t.Errorf("expected 2 doctors in result, got %d", len(doctors))
	}
}

func TestService_CreatePrescription(t *testing.T) {
	svc := service.New(
		&mockDoctorRepo{},
		&mockSubscriptionRepo{},
		&mockMessageRepo{},
		&mockPrescriptionRepo{},
		&mockConsultationRepo{},
		&mockModificationRepo{},
	)

	p := &model.Prescription{
		UserID:           "user-1",
		DoctorID:         "doc-1",
		PrescriptionType: "recommendation",
		Title:            "Rest for 3 days",
		Priority:         "normal",
	}
	result, err := svc.CreatePrescription(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", result.Status)
	}
	if result.ID == "" {
		t.Error("expected prescription ID to be set")
	}
}

func TestService_ScheduleConsultation_NoSubscription(t *testing.T) {
	subRepo := &mockSubscriptionRepo{activeCheck: false}
	svc := service.New(
		&mockDoctorRepo{},
		subRepo,
		&mockMessageRepo{},
		&mockPrescriptionRepo{},
		&mockConsultationRepo{},
		&mockModificationRepo{},
	)

	c := &model.Consultation{
		UserID:      "user-1",
		DoctorID:    "doc-1",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	}
	_, err := svc.ScheduleConsultation(context.Background(), c)
	if err == nil {
		t.Fatal("expected error when no active subscription")
	}
}

func TestSubscriptionPlanFunctions(t *testing.T) {
	tests := []struct {
		name      string
		plan      string
		wantPrice float64
		wantOk    bool
	}{
		{"monthly", "monthly", 999.0, true},
		{"quarterly", "quarterly", 2499.0, true},
		{"yearly", "yearly", 7999.0, true},
		{"invalid", "invalid", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, ok := model.SubscriptionPlanPrice(tt.plan)
			if ok != tt.wantOk {
				t.Errorf("price ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && price != tt.wantPrice {
				t.Errorf("price = %v, want %v", price, tt.wantPrice)
			}
		})
	}

	// Duration tests
	durTests := []struct {
		name   string
		plan   string
		wantOk bool
	}{
		{"monthly", "monthly", true},
		{"quarterly", "quarterly", true},
		{"yearly", "yearly", true},
		{"invalid", "invalid", false},
	}

	for _, tt := range durTests {
		t.Run(tt.name, func(t *testing.T) {
			dur, ok := model.SubscriptionPlanDuration(tt.plan)
			if ok != tt.wantOk {
				t.Errorf("duration ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && dur <= 0 {
				t.Errorf("expected positive duration, got %v", dur)
			}
		})
	}
}
