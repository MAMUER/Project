package postgres

import (
	"context"
	"database/sql"

	"github.com/MAMUER/project/internal/apperrors"
	"github.com/MAMUER/project/internal/domain/port"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, ev *port.EmailVerification) error
	GetValidToken(ctx context.Context, token string) (*port.EmailVerification, error)
	GetByUserID(ctx context.Context, userID string) (*port.EmailVerification, error)
	MarkUsed(ctx context.Context, token string) error
	MarkUserEmailVerified(ctx context.Context, userID string) error
}

type emailVerificationRepository struct {
	db *sql.DB
}

func NewEmailVerificationRepository(db *sql.DB) port.EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ctx context.Context, ev *port.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (user_id, email, email_hash, token, used, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		ev.UserID, ev.Email, ev.EmailHash, ev.Token, ev.Used, ev.ExpiresAt, ev.CreatedAt,
	)
	if err != nil {
		return apperrors.Internal("failed to create email verification", err)
	}
	return nil
}

func (r *emailVerificationRepository) GetValidToken(ctx context.Context, token string) (*port.EmailVerification, error) {
	query := `
		SELECT id, user_id, email, email_hash, token, used, expires_at, created_at
		FROM email_verifications
		WHERE token = $1 AND used = false AND expires_at > NOW()
	`
	ev := &port.EmailVerification{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&ev.ID, &ev.UserID, &ev.Email, &ev.EmailHash, &ev.Token, &ev.Used, &ev.ExpiresAt, &ev.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("email verification token not found")
		}
		return nil, apperrors.Internal("failed to get email verification", err)
	}
	return ev, nil
}

func (r *emailVerificationRepository) GetByUserID(ctx context.Context, userID string) (*port.EmailVerification, error) {
	query := `
		SELECT id, user_id, email, email_hash, token, used, expires_at, created_at
		FROM email_verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	ev := &port.EmailVerification{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&ev.ID, &ev.UserID, &ev.Email, &ev.EmailHash, &ev.Token, &ev.Used, &ev.ExpiresAt, &ev.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("email verification not found")
		}
		return nil, apperrors.Internal("failed to get email verification by user id", err)
	}
	return ev, nil
}

func (r *emailVerificationRepository) MarkUsed(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE email_verifications SET used = true WHERE token = $1`, token)
	if err != nil {
		return apperrors.Internal("failed to mark email verification as used", err)
	}
	return nil
}

func (r *emailVerificationRepository) MarkUserEmailVerified(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET email_verified = true WHERE id = $1`, userID)
	if err != nil {
		return apperrors.Internal("failed to mark user email as verified", err)
	}
	return nil
}
