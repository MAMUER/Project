package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	pb "github.com/MAMUER/Project/api/gen/user"
	"github.com/MAMUER/Project/internal/auth"
	"github.com/MAMUER/Project/internal/db"
	"github.com/MAMUER/Project/internal/logger"
	"github.com/MAMUER/Project/pkg/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userServer struct {
	pb.UnimplementedUserServiceServer
	db     *sql.DB
	log    *logger.Logger
	secret string
}

// sanitizeString очищает строку от потенциально опасных символов
func sanitizeString(s string) string {
	// Удаляем HTML-теги и потенциально опасные символы
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `\`, `\\`)
	return s
}

// isValidEmail проверяет формат email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// safeJSONArray безопасно создает JSON массив из строки
func safeJSONArray(items []string) (string, error) {
	if len(items) == 0 {
		return "[]", nil
	}
	// Санитизируем каждый элемент
	sanitized := make([]string, len(items))
	for i, item := range items {
		sanitized[i] = sanitizeString(item)
	}
	jsonBytes, err := json.Marshal(sanitized)
	if err != nil {
		return "[]", fmt.Errorf("failed to marshal JSON array: %w", err)
	}
	return string(jsonBytes), nil
}

// validateRegisterRequest проверяет данные регистрации
func validateRegisterRequest(req *pb.RegisterRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if !isValidEmail(req.Email) {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}
	if req.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if len(req.Password) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	if req.FullName == "" {
		return status.Error(codes.InvalidArgument, "full name is required")
	}
	if req.Role == "" {
		return status.Error(codes.InvalidArgument, "role is required")
	}
	validRoles := map[string]bool{"client": true, "admin": true, "doctor": true}
	if !validRoles[req.Role] {
		return status.Error(codes.InvalidArgument, "invalid role, must be client, admin, or doctor")
	}
	return nil
}

// validateLoginRequest проверяет данные для входа
func validateLoginRequest(req *pb.LoginRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

// validateProfileUpdate проверяет данные обновления профиля
func validateProfileUpdate(req *pb.UpdateProfileRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Age != nil && (*req.Age < 0 || *req.Age > 150) {
		return status.Error(codes.InvalidArgument, "age must be between 0 and 150")
	}
	if req.HeightCm != nil && (*req.HeightCm < 50 || *req.HeightCm > 300) {
		return status.Error(codes.InvalidArgument, "height_cm must be between 50 and 300")
	}
	if req.WeightKg != nil && (*req.WeightKg < 1 || *req.WeightKg > 500) {
		return status.Error(codes.InvalidArgument, "weight_kg must be between 1 and 500")
	}
	validFitnessLevels := map[string]bool{"": true, "beginner": true, "intermediate": true, "advanced": true}
	if req.FitnessLevel != nil && !validFitnessLevels[*req.FitnessLevel] {
		return status.Error(codes.InvalidArgument, "fitness_level must be beginner, intermediate, or advanced")
	}
	validGenders := map[string]bool{"": true, "male": true, "female": true, "other": true}
	if req.Gender != nil && !validGenders[*req.Gender] {
		return status.Error(codes.InvalidArgument, "gender must be male, female, or other")
	}
	return nil
}

func (s *userServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.log.Info("Register request", zap.String("email", req.Email))

	// Валидация входных данных
	if err := validateRegisterRequest(req); err != nil {
		s.log.Warn("Invalid register request", zap.Error(err))
		return nil, err
	}

	// Санитизируем входные данные
	email := sanitizeString(req.Email)
	fullName := sanitizeString(req.FullName)

	// Проверка существования пользователя
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		s.log.Error("Database error checking user existence", zap.Error(err))
		return nil, status.Error(codes.Internal, "database error")
	}
	if exists {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	// Хэширование пароля
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("Failed to hash password", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	// Создание пользователя
	userID := uuid.New().String()
	_, err = s.db.ExecContext(ctx, `
        INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    `, userID, email, string(hashed), fullName, req.Role)
	if err != nil {
		s.log.Error("Failed to create user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Создание пустого профиля
	_, err = s.db.ExecContext(ctx, `
        INSERT INTO user_profiles (user_id) VALUES ($1)
    `, userID)
	if err != nil {
		// Не критично, но логируем
		s.log.Warn("Failed to create user profile", zap.Error(err))
	}

	return &pb.RegisterResponse{
		UserId:  userID,
		Message: "user created successfully",
	}, nil
}

func (s *userServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.log.Info("Login request", zap.String("email", req.Email))

	// Валидация входных данных
	if err := validateLoginRequest(req); err != nil {
		s.log.Warn("Invalid login request", zap.Error(err))
		return nil, err
	}

	var user models.User
	err := s.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, role FROM users WHERE email = $1
    `, req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err == sql.ErrNoRows {
		// Возвращаем Unauthenticated вместо NotFound для безопасности
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	if err != nil {
		s.log.Error("Database error during login", zap.Error(err))
		return nil, status.Error(codes.Internal, "database error")
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.log.Info("Invalid login attempt", zap.String("email", req.Email))
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Генерация JWT
	token, err := auth.GenerateJWT(user.ID, user.Email, user.Role, s.secret, 24)
	if err != nil {
		s.log.Error("Failed to generate JWT", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   24 * 3600,
		UserId:      user.ID,
		Role:        user.Role,
	}, nil
}

func (s *userServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserProfile, error) {
	var profile pb.UserProfile
	var age sql.NullInt32
	var gender sql.NullString
	var heightCm sql.NullInt32
	var weightKg sql.NullFloat64
	var fitnessLevel sql.NullString
	var goals, contraindications string

	err := s.db.QueryRow(`
        SELECT u.id, u.email, u.full_name, u.role,
               p.age, p.gender, p.height_cm, p.weight_kg, p.fitness_level,
               COALESCE(p.goals::text, '[]') as goals,
               COALESCE(p.contraindications::text, '[]') as contraindications,
               u.created_at, u.updated_at
        FROM users u
        LEFT JOIN user_profiles p ON u.id = p.user_id
        WHERE u.id = $1
    `, req.UserId).Scan(
		&profile.UserId, &profile.Email, &profile.FullName, &profile.Role,
		&age, &gender, &heightCm, &weightKg, &fitnessLevel,
		&goals, &contraindications,
		&profile.CreatedAt, &profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	if err != nil {
		s.log.Error("Database error getting profile", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Error(codes.Internal, "database error")
	}

	if age.Valid {
		profile.Age = age.Int32
	}
	if gender.Valid {
		profile.Gender = gender.String
	}
	if heightCm.Valid {
		profile.HeightCm = heightCm.Int32
	}
	if weightKg.Valid {
		profile.WeightKg = weightKg.Float64
	}
	if fitnessLevel.Valid {
		profile.FitnessLevel = fitnessLevel.String
	}

	return &profile, nil
}

func (s *userServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UserProfile, error) {
	// Валидация входных данных
	if err := validateProfileUpdate(req); err != nil {
		s.log.Warn("Invalid profile update request", zap.Error(err))
		return nil, err
	}

	// Безопасное создание JSON массивов
	goalsJSON, err := safeJSONArray(req.Goals)
	if err != nil {
		s.log.Error("Failed to marshal goals", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to process goals")
	}
	contraindicationsJSON, err := safeJSONArray(req.Contraindications)
	if err != nil {
		s.log.Error("Failed to marshal contraindications", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to process contraindications")
	}

	query := `
        INSERT INTO user_profiles (user_id, age, gender, height_cm, weight_kg, fitness_level, goals, contraindications, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
        ON CONFLICT (user_id) DO UPDATE SET
            age = COALESCE(EXCLUDED.age, user_profiles.age),
            gender = COALESCE(EXCLUDED.gender, user_profiles.gender),
            height_cm = COALESCE(EXCLUDED.height_cm, user_profiles.height_cm),
            weight_kg = COALESCE(EXCLUDED.weight_kg, user_profiles.weight_kg),
            fitness_level = COALESCE(EXCLUDED.fitness_level, user_profiles.fitness_level),
            goals = COALESCE(EXCLUDED.goals, user_profiles.goals),
            contraindications = COALESCE(EXCLUDED.contraindications, user_profiles.contraindications),
            updated_at = NOW()
    `

	_, err = s.db.ExecContext(ctx, query,
		req.UserId,
		req.Age, req.Gender, req.HeightCm, req.WeightKg, req.FitnessLevel,
		goalsJSON, contraindicationsJSON,
	)
	if err != nil {
		s.log.Error("Failed to update profile", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Error(codes.Internal, "failed to update profile")
	}

	// Возвращаем обновленный профиль
	return s.GetProfile(ctx, &pb.GetProfileRequest{UserId: req.UserId})
}

func (s *userServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Валидация параметров
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.PageSize <= 0 {
		return nil, status.Error(codes.InvalidArgument, "page_size must be greater than 0")
	}
	if req.Page < 0 {
		return nil, status.Error(codes.InvalidArgument, "page must be non-negative")
	}

	offset := req.Page * req.PageSize
	rows, err := s.db.QueryContext(ctx, `
        SELECT u.id, u.email, u.full_name, u.role, u.created_at, u.updated_at
        FROM users u
        WHERE ($1 = '' OR u.role = $1)
        ORDER BY u.created_at DESC
        LIMIT $2 OFFSET $3
    `, req.Role, req.PageSize, offset)
	if err != nil {
		s.log.Error("Failed to list users", zap.Error(err))
		return nil, status.Error(codes.Internal, "database error")
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.log.Warn("Failed to close rows", zap.Error(closeErr))
		}
	}()

	var users []*pb.UserProfile
	for rows.Next() {
		var user pb.UserProfile
		if err := rows.Scan(&user.UserId, &user.Email, &user.FullName, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			s.log.Error("Failed to scan user", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to read user data")
		}
		users = append(users, &user)
	}

	// Проверяем ошибку итерации
	if err := rows.Err(); err != nil {
		s.log.Error("Row iteration error", zap.Error(err))
		return nil, status.Error(codes.Internal, "error reading users")
	}

	var total int32
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE ($1 = '' OR role = $1)", req.Role).Scan(&total)
	if err != nil {
		s.log.Warn("Failed to count users", zap.Error(err))
		// Не блокируем ответ, просто логируем
	}

	return &pb.ListUsersResponse{
		Users: users,
		Total: total,
	}, nil
}

func main() {
	log := logger.New("user-service")
	defer log.Sync() //nolint:errcheck

	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "50051"
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

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
		log.Warn("Using default JWT secret", zap.String("secret", secret))
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &userServer{
		db:     database,
		log:    log,
		secret: secret,
	})

	log.Info("User service starting", zap.String("port", port))
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve", zap.Error(err))
	}
}
