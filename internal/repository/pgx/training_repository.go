package pgx

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/MAMUER/project/internal/apperrors"
	"github.com/MAMUER/project/internal/domain/entity"
)

// TrainingRepositoryPGX implements training operations using database/sql for now.
// This is a bridge to allow gradual migration to pgx.
type TrainingRepositoryPGX struct {
	db *sql.DB
}

func NewTrainingRepositoryPGX(db *sql.DB) *TrainingRepositoryPGX {
	return &TrainingRepositoryPGX{db: db}
}

func (r *TrainingRepositoryPGX) CreatePlan(ctx context.Context, plan *entity.TrainingPlan) (*entity.TrainingPlan, error) {
	planDataJSON, err := json.Marshal(plan.PlanData)
	if err != nil {
		return nil, apperrors.Internal("failed to marshal plan data", err)
	}

	query := `
		INSERT INTO training_plans (id, user_id, classification, duration_weeks, available_days, plan_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err = r.db.QueryRowContext(ctx, query,
		plan.ID, plan.UserID, plan.Classification, plan.DurationWeeks, plan.AvailableDays,
		planDataJSON, plan.CreatedAt, plan.UpdatedAt,
	).Scan(&plan.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to create training plan", err)
	}
	return plan, nil
}

func (r *TrainingRepositoryPGX) GetPlan(ctx context.Context, userID, planID string) (*entity.TrainingPlan, error) {
	query := `
		SELECT id, user_id, classification, duration_weeks, available_days, plan_data, created_at, updated_at
		FROM training_plans WHERE id = $1 AND user_id = $2
	`
	plan := &entity.TrainingPlan{}
	var planDataJSON []byte

	err := r.db.QueryRowContext(ctx, query, planID, userID).Scan(
		&plan.ID, &plan.UserID, &plan.Classification, &plan.DurationWeeks,
		&plan.AvailableDays, &planDataJSON, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("training plan not found")
		}
		return nil, apperrors.Internal("failed to get training plan", err)
	}

	if len(planDataJSON) > 0 {
		if err := json.Unmarshal(planDataJSON, &plan.PlanData); err != nil {
			return nil, apperrors.Internal("failed to unmarshal plan data", err)
		}
	}

	return plan, nil
}

func (r *TrainingRepositoryPGX) ListPlans(ctx context.Context, userID string, page, pageSize int) ([]*entity.TrainingPlan, int, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT id, user_id, classification, duration_weeks, available_days, plan_data, created_at, updated_at
		FROM training_plans WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to list training plans", err)
	}
	defer rows.Close()

	var plans []*entity.TrainingPlan
	for rows.Next() {
		plan := &entity.TrainingPlan{}
		var planDataJSON []byte
		if err := rows.Scan(
			&plan.ID, &plan.UserID, &plan.Classification, &plan.DurationWeeks,
			&plan.AvailableDays, &planDataJSON, &plan.CreatedAt, &plan.UpdatedAt,
		); err != nil {
			return nil, 0, apperrors.Internal("failed to scan training plan", err)
		}
		if len(planDataJSON) > 0 {
			if err := json.Unmarshal(planDataJSON, &plan.PlanData); err != nil {
				return nil, 0, apperrors.Internal("failed to unmarshal plan data", err)
			}
		}
		plans = append(plans, plan)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM training_plans WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, apperrors.Internal("failed to count training plans", err)
	}

	return plans, total, nil
}

func (r *TrainingRepositoryPGX) CompleteWorkout(ctx context.Context, userID, planID string) error {
	query := `
		UPDATE training_plans SET updated_at = $1 WHERE id = $2 AND user_id = $3
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), planID, userID)
	if err != nil {
		return apperrors.Internal("failed to complete workout", err)
	}
	return nil
}

func (r *TrainingRepositoryPGX) GetProgress(ctx context.Context, userID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_plans,
			COUNT(CASE WHEN updated_at > created_at THEN 1 END) as completed_workouts
		FROM training_plans WHERE user_id = $1
	`
	var totalPlans, completedWorkouts int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&totalPlans, &completedWorkouts)
	if err != nil {
		return nil, apperrors.Internal("failed to get progress", err)
	}

	return map[string]interface{}{
		"total_plans":        totalPlans,
		"completed_workouts": completedWorkouts,
		"completion_rate":    float64(completedWorkouts) / float64(totalPlans) * 100,
	}, nil
}

func (r *TrainingRepositoryPGX) GetAchievements(ctx context.Context, userID string) ([]*entity.Achievement, error) {
	query := `
		SELECT id, user_id, type, title, description, earned_at
		FROM achievements WHERE user_id = $1 ORDER BY earned_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to get achievements", err)
	}
	defer rows.Close()

	var achievements []*entity.Achievement
	for rows.Next() {
		achievement := &entity.Achievement{}
		if err := rows.Scan(
			&achievement.ID, &achievement.UserID, &achievement.Type, &achievement.Title, &achievement.Description, &achievement.EarnedAt,
		); err != nil {
			return nil, apperrors.Internal("failed to scan achievement", err)
		}
		achievements = append(achievements, achievement)
	}
	return achievements, nil
}

func (r *TrainingRepositoryPGX) CreateAchievement(ctx context.Context, achievement *entity.Achievement) (*entity.Achievement, error) {
	query := `
		INSERT INTO achievements (id, user_id, type, title, description, earned_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		achievement.ID, achievement.UserID, achievement.Type, achievement.Title, achievement.Description, achievement.EarnedAt,
	).Scan(&achievement.ID)
	if err != nil {
		return nil, apperrors.Internal("failed to create achievement", err)
	}
	return achievement, nil
}

func (r *TrainingRepositoryPGX) DeletePlan(ctx context.Context, userID, planID string) error {
	query := `DELETE FROM training_plans WHERE id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, planID, userID)
	if err != nil {
		return apperrors.Internal("failed to delete plan", err)
	}
	return nil
}

func (r *TrainingRepositoryPGX) UpdatePlan(ctx context.Context, plan *entity.TrainingPlan) (*entity.TrainingPlan, error) {
	query := `
		UPDATE training_plans SET classification = $1, duration_weeks = $2, available_days = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
		RETURNING id, user_id, classification, duration_weeks, available_days, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		plan.Classification, plan.DurationWeeks, plan.AvailableDays, time.Now(),
		plan.ID, plan.UserID,
	).Scan(&plan.ID, &plan.UserID, &plan.Classification, &plan.DurationWeeks, &plan.AvailableDays, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("training plan not found")
		}
		return nil, apperrors.Internal("failed to update training plan", err)
	}
	return plan, nil
}
