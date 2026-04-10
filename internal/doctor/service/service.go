// Package service implements the application layer (use cases) for the doctor service.
// It orchestrates domain logic using port interfaces, with no direct infrastructure dependencies.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/MAMUER/Project/internal/doctor/model"
	"github.com/MAMUER/Project/internal/doctor/port"
	"github.com/google/uuid"
)

// Service implements port.DoctorService using the repository ports.
type Service struct {
	doctors       port.DoctorRepository
	subscriptions port.SubscriptionRepository
	messages      port.MessageRepository
	prescriptions port.PrescriptionRepository
	consultations port.ConsultationRepository
	modifications port.TrainingModificationRepository
}

// New creates a new doctor service with the given repository implementations.
func New(
	doctors port.DoctorRepository,
	subscriptions port.SubscriptionRepository,
	messages port.MessageRepository,
	prescriptions port.PrescriptionRepository,
	consultations port.ConsultationRepository,
	modifications port.TrainingModificationRepository,
) *Service {
	return &Service{
		doctors:       doctors,
		subscriptions: subscriptions,
		messages:      messages,
		prescriptions: prescriptions,
		consultations: consultations,
		modifications: modifications,
	}
}

// ===== Doctors =====

func (s *Service) ListDoctors(ctx context.Context, specialty string, page, pageSize int32) ([]model.Doctor, int64, error) {
	return s.doctors.List(ctx, specialty, page, pageSize)
}

func (s *Service) GetDoctor(ctx context.Context, id string) (*model.Doctor, error) {
	return s.doctors.GetByID(ctx, id)
}

// ===== Subscriptions =====

func (s *Service) Subscribe(ctx context.Context, userID, doctorID, planType string) (*model.Subscription, error) {
	// Validate plan type
	duration, ok := model.SubscriptionPlanDuration(planType)
	if !ok {
		return nil, errors.New("invalid plan type")
	}

	price, _ := model.SubscriptionPlanPrice(planType)

	subscription := &model.Subscription{
		ID:        uuid.New().String(),
		UserID:    userID,
		DoctorID:  doctorID,
		PlanType:  planType,
		StartsAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
		IsActive:  true,
		Price:     price,
		CreatedAt: time.Now(),
	}

	if err := s.subscriptions.Create(ctx, subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

func (s *Service) GetSubscription(ctx context.Context, userID, doctorID string) (*model.Subscription, error) {
	return s.subscriptions.GetByUserAndDoctor(ctx, userID, doctorID)
}

func (s *Service) CancelSubscription(ctx context.Context, userID, doctorID string) error {
	return s.subscriptions.Cancel(ctx, userID, doctorID)
}

// ===== Messages =====

func (s *Service) SendMessage(ctx context.Context, m *model.Message) (*model.Message, error) {
	// Verify active subscription
	active, err := s.subscriptions.IsActive(ctx, m.UserID, m.DoctorID)
	if err != nil || !active {
		return nil, errors.New("active subscription required")
	}

	m.ID = uuid.New().String()
	m.IsRead = false
	m.CreatedAt = time.Now()

	return s.messages.Create(ctx, m)
}

func (s *Service) GetChatHistory(ctx context.Context, userID, doctorID string, page, pageSize int32) ([]model.Message, error) {
	return s.messages.GetChatHistory(ctx, userID, doctorID, page, pageSize)
}

func (s *Service) MarkMessagesRead(ctx context.Context, userID, doctorID string) (int64, error) {
	return s.messages.MarkAsRead(ctx, userID, doctorID)
}

func (s *Service) GetUnreadCount(ctx context.Context, userID string) (map[string]int64, int64, error) {
	return s.messages.GetUnreadCount(ctx, userID)
}

// ===== Prescriptions =====

func (s *Service) CreatePrescription(ctx context.Context, p *model.Prescription) (*model.Prescription, error) {
	p.ID = uuid.New().String()
	p.Status = "active"
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	return s.prescriptions.Create(ctx, p)
}

func (s *Service) GetPrescriptions(ctx context.Context, userID, statusFilter string) ([]model.Prescription, error) {
	return s.prescriptions.GetByUser(ctx, userID, statusFilter)
}

func (s *Service) UpdatePrescriptionStatus(ctx context.Context, id, status string) error {
	return s.prescriptions.UpdateStatus(ctx, id, status)
}

// ===== Training Modifications =====

func (s *Service) CreateTrainingModification(ctx context.Context, tm *model.TrainingModification) error {
	tm.ID = uuid.New().String()
	tm.CreatedAt = time.Now()
	return s.modifications.Create(ctx, tm)
}

func (s *Service) GetTrainingModifications(ctx context.Context, userID, trainingPlanID string) ([]model.TrainingModification, error) {
	return s.modifications.GetByUser(ctx, userID, trainingPlanID)
}

// ===== Consultations =====

func (s *Service) ScheduleConsultation(ctx context.Context, c *model.Consultation) (*model.Consultation, error) {
	// Verify active subscription
	active, err := s.subscriptions.IsActive(ctx, c.UserID, c.DoctorID)
	if err != nil || !active {
		return nil, errors.New("active subscription required")
	}

	c.ID = uuid.New().String()
	c.Status = "scheduled"
	c.CreatedAt = time.Now()

	return s.consultations.Create(ctx, c)
}

func (s *Service) GetConsultations(ctx context.Context, userID, statusFilter string) ([]model.Consultation, error) {
	return s.consultations.GetByUser(ctx, userID, statusFilter)
}

func (s *Service) CompleteConsultation(ctx context.Context, id, notes string) error {
	return s.consultations.Complete(ctx, id, notes)
}
