// Package port defines the ports (interfaces) for the doctor service.
// These interfaces form the boundary of the hexagon — the application
// depends on these abstractions, not on concrete implementations.
package port

import (
	"context"

	"github.com/MAMUER/Project/internal/doctor/model"
)

// DoctorRepository is an output port for doctor persistence.
type DoctorRepository interface {
	List(ctx context.Context, specialty string, page, pageSize int32) ([]model.Doctor, int64, error)
	GetByID(ctx context.Context, id string) (*model.Doctor, error)
	GetByUserID(ctx context.Context, userID string) (*model.Doctor, error)
	Create(ctx context.Context, d *model.Doctor) error
	Update(ctx context.Context, d *model.Doctor) error
}

// SubscriptionRepository is an output port for subscription persistence.
type SubscriptionRepository interface {
	Create(ctx context.Context, s *model.Subscription) error
	GetByUserAndDoctor(ctx context.Context, userID, doctorID string) (*model.Subscription, error)
	IsActive(ctx context.Context, userID, doctorID string) (bool, error)
	Cancel(ctx context.Context, userID, doctorID string) error
}

// MessageRepository is an output port for chat message persistence.
type MessageRepository interface {
	Create(ctx context.Context, m *model.Message) (*model.Message, error)
	GetChatHistory(ctx context.Context, userID, doctorID string, page, pageSize int32) ([]model.Message, error)
	MarkAsRead(ctx context.Context, userID, doctorID string) (int64, error)
	GetUnreadCount(ctx context.Context, userID string) (map[string]int64, int64, error)
}

// PrescriptionRepository is an output port for prescription persistence.
type PrescriptionRepository interface {
	Create(ctx context.Context, p *model.Prescription) (*model.Prescription, error)
	GetByUser(ctx context.Context, userID, statusFilter string) ([]model.Prescription, error)
	UpdateStatus(ctx context.Context, id, status string) error
}

// ConsultationRepository is an output port for consultation persistence.
type ConsultationRepository interface {
	Create(ctx context.Context, c *model.Consultation) (*model.Consultation, error)
	GetByUser(ctx context.Context, userID, statusFilter string) ([]model.Consultation, error)
	Complete(ctx context.Context, id, notes string) error
}

// TrainingModificationRepository is an output port for training modification persistence.
type TrainingModificationRepository interface {
	Create(ctx context.Context, tm *model.TrainingModification) error
	GetByUser(ctx context.Context, userID, trainingPlanID string) ([]model.TrainingModification, error)
}

// DoctorService defines the input port (use cases) for the doctor service.
// This is what adapters (gRPC, HTTP) call into.
type DoctorService interface {
	// Doctors
	ListDoctors(ctx context.Context, specialty string, page, pageSize int32) ([]model.Doctor, int64, error)
	GetDoctor(ctx context.Context, id string) (*model.Doctor, error)

	// Subscriptions
	Subscribe(ctx context.Context, userID, doctorID, planType string) (*model.Subscription, error)
	GetSubscription(ctx context.Context, userID, doctorID string) (*model.Subscription, error)
	CancelSubscription(ctx context.Context, userID, doctorID string) error

	// Messages
	SendMessage(ctx context.Context, m *model.Message) (*model.Message, error)
	GetChatHistory(ctx context.Context, userID, doctorID string, page, pageSize int32) ([]model.Message, error)
	MarkMessagesRead(ctx context.Context, userID, doctorID string) (int64, error)
	GetUnreadCount(ctx context.Context, userID string) (map[string]int64, int64, error)

	// Prescriptions
	CreatePrescription(ctx context.Context, p *model.Prescription) (*model.Prescription, error)
	GetPrescriptions(ctx context.Context, userID, statusFilter string) ([]model.Prescription, error)
	UpdatePrescriptionStatus(ctx context.Context, id, status string) error

	// Training Modifications
	CreateTrainingModification(ctx context.Context, tm *model.TrainingModification) error
	GetTrainingModifications(ctx context.Context, userID, trainingPlanID string) ([]model.TrainingModification, error)

	// Consultations
	ScheduleConsultation(ctx context.Context, c *model.Consultation) (*model.Consultation, error)
	GetConsultations(ctx context.Context, userID, statusFilter string) ([]model.Consultation, error)
	CompleteConsultation(ctx context.Context, id, notes string) error
}
