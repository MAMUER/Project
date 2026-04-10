// Package model defines the domain models for the doctor service.
// These are pure domain types with no dependencies on infrastructure.
package model

import "time"

// Doctor represents a medical professional in the system.
type Doctor struct {
	ID            string
	UserID        *string
	Specialty     string
	LicenseNumber string
	Phone         string
	Bio           string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Subscription represents a user's subscription to a doctor.
type Subscription struct {
	ID        string
	UserID    string
	DoctorID  string
	PlanType  string
	StartsAt  time.Time
	ExpiresAt time.Time
	IsActive  bool
	Price     float64
	CreatedAt time.Time
}

// Message represents a chat message between user and doctor.
type Message struct {
	ID             string
	UserID         string
	DoctorID       string
	SenderUserID   *string
	SenderDoctorID *string
	Message        string
	MessageType    string
	IsRead         bool
	CreatedAt      time.Time
}

// Prescription represents a doctor's prescription for a patient.
type Prescription struct {
	ID               string
	ConsultationID   *string
	UserID           string
	DoctorID         string
	PrescriptionType string
	Title            string
	Description      string
	Priority         string
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Consultation represents a scheduled consultation session.
type Consultation struct {
	ID          string
	UserID      string
	DoctorID    string
	Status      string
	ScheduledAt time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
	Notes       string
	CreatedAt   time.Time
}

// TrainingModification represents a change to a training plan made by a doctor.
type TrainingModification struct {
	ID               string
	DoctorID         string
	TrainingPlanID   string
	ModificationType string
	OldValue         string
	NewValue         string
	Reason           string
	CreatedAt        time.Time
}

// DoctorStats represents aggregated statistics for a doctor.
type DoctorStats struct {
	DoctorID          string
	ConsultationCount int64
	AvgRating         float64
}

// SubscriptionPlanPrice returns the price for a given plan type.
func SubscriptionPlanPrice(planType string) (float64, bool) {
	prices := map[string]float64{
		"monthly":   999.0,
		"quarterly": 2499.0,
		"yearly":    7999.0,
	}
	price, ok := prices[planType]
	return price, ok
}

// SubscriptionPlanDuration returns the duration for a given plan type.
func SubscriptionPlanDuration(planType string) (time.Duration, bool) {
	durations := map[string]time.Duration{
		"monthly":   30 * 24 * time.Hour,
		"quarterly": 90 * 24 * time.Hour,
		"yearly":    365 * 24 * time.Hour,
	}
	d, ok := durations[planType]
	return d, ok
}
