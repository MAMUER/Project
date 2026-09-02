package postgres

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MAMUER/project/internal/apperrors"
	"github.com/MAMUER/project/internal/db"
	"github.com/MAMUER/project/internal/domain/entity"
	"github.com/lib/pq"
)

// PgsodiumUserRepository handles user operations with pgsodium encryption.
type PgsodiumUserRepository struct {
	db *sql.DB
}

func NewPgsodiumUserRepository(db *sql.DB) *PgsodiumUserRepository {
	return &PgsodiumUserRepository{db: db}
}

func (r *PgsodiumUserRepository) Create(ctx context.Context, user *entity.User) error {
	emailNonce, err := db.GenerateNonce()
	if err != nil {
		return apperrors.Internal("failed to generate nonce", err)
	}
	fullNameNonce, err := db.GenerateNonce()
	if err != nil {
		return apperrors.Internal("failed to generate nonce", err)
	}

	emailHash := db.BlindIndex(user.Email)
	fullNameHash := db.BlindIndex(user.FullName)

	query, args := buildUserInsertQuery(userInsertData{
		userID: user.ID, email: user.Email, emailHash: emailHash, passwordHash: user.PasswordHash,
		fullName: user.FullName, emailNonce: emailNonce, fullNameNonce: fullNameNonce,
		fullNameHash: fullNameHash, role: user.Role,
	})
	if _, execErr := r.db.ExecContext(ctx, query, args...); execErr != nil {
		return apperrors.Internal("failed to create user", execErr)
	}
	return nil
}

func (r *PgsodiumUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	var emailHash, fullNameHash string
	err := r.db.QueryRowContext(ctx, "SELECT id, email_hash, password_hash, full_name_hash, role, email_verified, created_at, updated_at FROM users WHERE id = $1", id).Scan(
		&user.ID, &emailHash, &user.PasswordHash, &fullNameHash,
		&user.Role, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NotFound("user not found")
	}
	if err != nil {
		return nil, apperrors.Internal("failed to get user", err)
	}

	if db.PgsodiumKeyID() > 0 {
		var email, fullName string
		emailQuery := strings.Builder{}
		emailQuery.WriteString("SELECT ")
		emailQuery.WriteString(db.PgsodiumDecryptParam("email_encrypted", "email_nonce", "email"))
		emailQuery.WriteString(" FROM users WHERE id = $1")
		if err := r.db.QueryRowContext(ctx, emailQuery.String(), user.ID).Scan(&email); err != nil {
			return nil, apperrors.Internal("failed to decrypt email", err)
		}
		fullNameQuery := strings.Builder{}
		fullNameQuery.WriteString("SELECT ")
		fullNameQuery.WriteString(db.PgsodiumDecryptParam("full_name_encrypted", "full_name_nonce", "full_name"))
		fullNameQuery.WriteString(" FROM users WHERE id = $1")
		if err := r.db.QueryRowContext(ctx, fullNameQuery.String(), user.ID).Scan(&fullName); err != nil {
			return nil, apperrors.Internal("failed to decrypt full name", err)
		}
		user.Email = email
		user.FullName = fullName
	}

	return &user, nil
}

func (r *PgsodiumUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	emailHash := db.BlindIndex(email)
	var user entity.User
	var fullNameHash string
	err := r.db.QueryRowContext(ctx, "SELECT id, email_hash, password_hash, full_name_hash, role, email_verified, created_at, updated_at FROM users WHERE email_hash = $1", emailHash).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &fullNameHash,
		&user.Role, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NotFound("user not found")
	}
	if err != nil {
		return nil, apperrors.Internal("failed to get user by email", err)
	}

	if db.PgsodiumKeyID() > 0 {
		var emailVal string
		emailQuery := strings.Builder{}
		emailQuery.WriteString("SELECT ")
		emailQuery.WriteString(db.PgsodiumDecryptParam("email_encrypted", "email_nonce", "email"))
		emailQuery.WriteString(" FROM users WHERE id = $1")
		if err := r.db.QueryRowContext(ctx, emailQuery.String(), user.ID).Scan(&emailVal); err != nil {
			return nil, apperrors.Internal("failed to decrypt email", err)
		}
		user.Email = emailVal
	}

	return &user, nil
}

func (r *PgsodiumUserRepository) Update(ctx context.Context, user *entity.User) error {
	fullNameNonce, err := db.GenerateNonce()
	if err != nil {
		return apperrors.Internal("failed to generate nonce", err)
	}
	fullNameHash := db.BlindIndex(user.FullName)

	query := strings.Builder{}
	query.WriteString("UPDATE users SET full_name_encrypted = ")
	query.WriteString(db.PgsodiumRandomEncryptParam(1, 2))
	query.WriteString(", full_name_nonce = $2, full_name_hash = $3, updated_at = $4 WHERE id = $5")
	_, execErr := r.db.ExecContext(ctx, query.String(), user.FullName, fullNameNonce, fullNameHash, time.Now(), user.ID)
	if execErr != nil {
		return apperrors.Internal("failed to update user", execErr)
	}
	return nil
}

func (r *PgsodiumUserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return apperrors.Internal("failed to delete user", err)
	}
	return nil
}

func (r *PgsodiumUserRepository) List(ctx context.Context, page, pageSize int) ([]*entity.User, error) {
	offset := (page - 1) * pageSize
	var listUsersQuery strings.Builder
	listUsersQuery.WriteString("SELECT u.id, ")
	listUsersQuery.WriteString(db.PgsodiumDecryptParam("u.email_encrypted", "u.email_nonce", "email"))
	listUsersQuery.WriteString(",\n               ")
	listUsersQuery.WriteString(db.PgsodiumDecryptParam("u.full_name_encrypted", "u.full_name_nonce", "full_name"))
	listUsersQuery.WriteString(",\n               u.role, u.email_verified, u.created_at, u.updated_at\n            FROM users u\n            WHERE 1=1")
	args := []interface{}{}
	argCount := 0
	argCount++
	listUsersQuery.WriteString(fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1))
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, listUsersQuery.String(), args...)
	if err != nil {
		return nil, apperrors.Internal("failed to list users", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, apperrors.Internal("failed to scan user", err)
		}
		users = append(users, user)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE 1=1").Scan(&total); err != nil {
		return nil, apperrors.Internal("failed to count users", err)
	}

	return users, nil
}

func (r *PgsodiumUserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, apperrors.Internal("failed to count users", err)
	}
	return count, nil
}

func (r *PgsodiumUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	emailHash := db.BlindIndex(email)
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email_hash = $1)", emailHash).Scan(&exists)
	if err != nil {
		return false, apperrors.Internal("failed to check user existence", err)
	}
	return exists, nil
}

func (r *PgsodiumUserRepository) ConfirmEmail(ctx context.Context, token string) error {
	_, _, err := r.ConfirmEmailWithPgsodium(ctx, token)
	return err
}

func (r *PgsodiumUserRepository) ConfirmEmailWithPgsodium(ctx context.Context, token string) (userID, email string, err error) {
	var used bool
	var expiresAt sql.NullTime
	confirmEmailQuery := strings.Builder{}
	confirmEmailQuery.WriteString("SELECT user_id, ")
	confirmEmailQuery.WriteString(db.PgsodiumDecryptParam("email_encrypted", "email_nonce", "email"))
	confirmEmailQuery.WriteString(", used, expires_at FROM email_verifications WHERE token = $1")
	err = r.db.QueryRowContext(ctx, confirmEmailQuery.String(), token).Scan(&userID, &email, &used, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", apperrors.InvalidArgument("invalid verification token")
	}
	if err != nil {
		return "", "", apperrors.Internal("database error checking verification token", err)
	}
	if used {
		return "", "", apperrors.InvalidArgument("verification token has already been used")
	}
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return "", "", apperrors.InvalidArgument("verification token has expired")
	}
	return userID, email, nil
}

func (r *PgsodiumUserRepository) CreateEmailVerification(ctx context.Context, userID, email, emailHash, verificationToken string, emailNonce, tokenNonce []byte) error {
	verificationQuery, verificationArgs := buildEmailVerificationInsertQuery(userID, email, emailHash, verificationToken, emailNonce, tokenNonce)
	if _, execErr := r.db.ExecContext(ctx, verificationQuery, verificationArgs...); execErr != nil {
		return apperrors.Internal("failed to create email verification record", execErr)
	}
	return nil
}

func (r *PgsodiumUserRepository) LoginWithPgsodium(ctx context.Context, email string) (*entity.User, error) {
	emailHash := db.BlindIndex(email)
	var emailConfirmed bool
	err := r.db.QueryRowContext(ctx, "SELECT email_confirmed FROM users WHERE email_hash = $1", emailHash).Scan(&emailConfirmed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NotFound("user not found")
	}
	if err != nil {
		return nil, apperrors.Internal("database error checking email confirmation", err)
	}
	if !emailConfirmed {
		return nil, apperrors.Unauthorized("email not confirmed")
	}

	var user entity.User
	loginQuery := strings.Builder{}
	loginQuery.WriteString("SELECT id, ")
	loginQuery.WriteString(db.PgsodiumDecryptParam("email_encrypted", "email_nonce", "email"))
	loginQuery.WriteString(", password_hash, role FROM users WHERE email_hash = $1")
	err = r.db.QueryRowContext(ctx, loginQuery.String(), emailHash).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.Unauthorized("invalid credentials")
	}
	if err != nil {
		return nil, apperrors.Internal("database error during login", err)
	}
	user.EmailVerified = emailConfirmed
	return &user, nil
}

func (r *PgsodiumUserRepository) CreateGoogleUser(ctx context.Context, googleSub, emailHash, emailVal, nickname, nicknameHash string) (userID, role string, emailConfirmed bool, err error) {
	nicknameNonce, err := db.GenerateNonce()
	if err != nil {
		return "", "", false, apperrors.Internal("failed to generate nonce", err)
	}
	emailNonce, err := db.GenerateNonce()
	if err != nil {
		return "", "", false, apperrors.Internal("failed to generate nonce", err)
	}

	userID = hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	query, args := buildGoogleUserInsertQuery(googleUserInsertData{
		userID: userID, emailVal: emailVal, emailHash: emailHash, emailNonce: emailNonce,
		nickname: nickname, nicknameNonce: nicknameNonce, nicknameHash: nicknameHash, googleSub: googleSub,
	})
	_, insertErr := r.db.ExecContext(ctx, query, args...)
	if insertErr != nil {
		var pqErr *pq.Error
		if errors.As(insertErr, &pqErr) && pqErr.Code == "23505" {
			return "", "", false, apperrors.Conflict("user already exists")
		}
		return "", "", false, apperrors.Internal("failed to create user", insertErr)
	}
	role = "client"
	emailConfirmed = true
	return userID, role, emailConfirmed, nil
}

type userInsertData struct {
	userID, email, emailHash, passwordHash, fullName string
	emailNonce, fullNameNonce                        []byte
	fullNameHash, role                               string
}

func buildUserInsertQuery(p userInsertData) (string, []interface{}) {
	var b strings.Builder
	b.WriteString("INSERT INTO users (id, email_encrypted, email_nonce, email_hash, password_hash, full_name_encrypted, full_name_nonce, full_name_hash, role, created_at, updated_at) ")
	b.WriteString("VALUES ($1, ")
	b.WriteString(db.PgsodiumRandomEncryptParam(2, 3))
	b.WriteString(", $4, ")
	b.WriteString("$5, $6, ")
	b.WriteString(db.PgsodiumRandomEncryptParam(7, 8))
	b.WriteString(", $9, ")
	b.WriteString("$10, ")
	b.WriteString("$11, NOW(), NOW())")
	args := []interface{}{p.userID, p.email, p.emailNonce, p.emailNonce, p.emailHash, p.passwordHash, p.fullName, p.fullNameNonce, p.fullNameNonce, p.fullNameHash, p.role}
	return b.String(), args
}

func buildEmailVerificationInsertQuery(userID, email, emailHash, verificationToken string, emailNonce, tokenNonce []byte) (string, []interface{}) {
	var b strings.Builder
	b.WriteString("INSERT INTO email_verifications (user_id, email_encrypted, email_nonce, email_hash, token, token_encrypted, token_nonce, expires_at, used) ")
	b.WriteString("VALUES ($1, ")
	b.WriteString(db.PgsodiumRandomEncryptParam(2, 3))
	b.WriteString(", $4, ")
	b.WriteString("$5, $6, ")
	b.WriteString(db.PgsodiumRandomEncryptParam(7, 8))
	b.WriteString(", $9, NOW() + INTERVAL '24 hours', false)")
	args := []interface{}{userID, email, emailNonce, emailNonce, emailHash, verificationToken, tokenNonce, tokenNonce}
	return b.String(), args
}

type googleUserInsertData struct {
	userID, emailVal, emailHash, nickname, googleSub string
	emailNonce, nicknameNonce                        []byte
	nicknameHash                                     string
}

func buildGoogleUserInsertQuery(p googleUserInsertData) (string, []interface{}) {
	var b strings.Builder
	b.WriteString("INSERT INTO users (id, email_encrypted, email_nonce, email_hash, password_hash, nickname_encrypted, nickname_nonce, nickname_hash, role, provider, external_id, email_confirmed, created_at, updated_at) ")
	b.WriteString("VALUES ($1, ")
	b.WriteString(db.PgsodiumRandomEncryptParam(2, 3))
	b.WriteString(", $4, ")
	b.WriteString("$5, $6, ")
	b.WriteString(db.PgsodiumRandomEncryptParam(7, 8))
	b.WriteString(", $9, ")
	b.WriteString("$10, ")
	b.WriteString("'client', 'google', $11, true, NOW(), NOW())")
	args := []interface{}{p.userID, p.emailVal, p.emailNonce, p.emailNonce, p.emailHash, "", p.nickname, p.nicknameNonce, p.nicknameNonce, p.nicknameHash, p.googleSub}
	return b.String(), args
}
