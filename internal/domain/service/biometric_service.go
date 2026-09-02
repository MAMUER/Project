package service

import (
	"context"
	"time"

	"github.com/MAMUER/project/internal/apperrors"
	"github.com/MAMUER/project/internal/domain/entity"
	"github.com/MAMUER/project/internal/domain/port"
)

type biometricService struct {
	biometrics port.BiometricRepository
}

func NewBiometricService(biometrics port.BiometricRepository) BiometricService {
	return &biometricService{biometrics: biometrics}
}

func (s *biometricService) AddRecord(ctx context.Context, record *entity.BiometricRecord) (*entity.BiometricRecord, error) {
	if record.UserID == "" || record.MetricType == "" {
		return nil, apperrors.Validation("user_id and metric_type are required")
	}
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	return s.biometrics.Create(ctx, record)
}

func (s *biometricService) BatchAddRecords(ctx context.Context, records []*entity.BiometricRecord) (int, error) {
	if len(records) == 0 {
		return 0, apperrors.Validation("records cannot be empty")
	}
	for _, rec := range records {
		if rec.UserID == "" || rec.MetricType == "" {
			return 0, apperrors.Validation("user_id and metric_type are required for all records")
		}
		if rec.Timestamp.IsZero() {
			rec.Timestamp = time.Now()
		}
	}
	return s.biometrics.BatchCreate(ctx, records)
}

func (s *biometricService) GetRecords(ctx context.Context, userID, metricType string, limit int) ([]*entity.BiometricRecord, error) {
	if userID == "" {
		return nil, apperrors.Validation("user_id is required")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	return s.biometrics.GetByUserID(ctx, userID, metricType, limit, 0)
}

func (s *biometricService) GetLatest(ctx context.Context, userID, metricType string) (*entity.BiometricRecord, error) {
	if userID == "" || metricType == "" {
		return nil, apperrors.Validation("user_id and metric_type are required")
	}
	records, err := s.biometrics.GetByUserID(ctx, userID, metricType, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, apperrors.NotFound("no records found")
	}
	return records[0], nil
}

func (s *biometricService) UpdateRecord(ctx context.Context, record *entity.BiometricRecord) (*entity.BiometricRecord, error) {
	if record.ID == "" {
		return nil, apperrors.Validation("id is required")
	}
	if record.Value < 0 {
		return nil, apperrors.Validation("value cannot be negative")
	}
	return s.biometrics.Create(ctx, record)
}

func (s *biometricService) DeleteRecord(ctx context.Context, id string) error {
	if id == "" {
		return apperrors.Validation("id is required")
	}
	return s.biometrics.Delete(ctx, id)
}
