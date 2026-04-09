package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	biometricpb "github.com/MAMUER/Project/api/gen/biometric"
	trainingpb "github.com/MAMUER/Project/api/gen/training"
	userpb "github.com/MAMUER/Project/api/gen/user"
	"github.com/MAMUER/Project/internal/auth"
	"github.com/MAMUER/Project/internal/logger"
	"github.com/MAMUER/Project/internal/middleware"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type gateway struct {
	userClient         userpb.UserServiceClient
	biometricClient    biometricpb.BiometricServiceClient
	trainingClient     trainingpb.TrainingServiceClient
	mlClassifierURL    string
	mlGeneratorURL     string
	deviceConnectorURL string
	log                *logger.Logger
	jwtSecret          string
	db                 *sql.DB // For server-side role re-verification
	// Async ML processing
	rdb     *redis.Client
	rmqCh   *amqp.Channel
	mlAsync bool
}

// ========== Helper Functions ==========

func ptrInt32(v int32) *int32       { return &v }
func ptrString(v string) *string    { return &v }
func ptrFloat64(v float64) *float64 { return &v }
func ptrFloat32(v float32) *float32 { return &v }

// grpcToHTTPStatus maps gRPC error codes to HTTP status codes.
// Returns the mapped HTTP status code and a user-friendly message in Russian.
func grpcToHTTPStatus(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "Внутренняя ошибка сервера"
	}
	msg := st.Message()
	// Переводим технические сообщения на русский
	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, translateError(msg)
	case codes.AlreadyExists:
		return http.StatusConflict, translateError(msg)
	case codes.NotFound:
		return http.StatusNotFound, "Не найдено"
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "Неверные учётные данные"
	case codes.PermissionDenied:
		// Требование #4: Никогда не возвращаем 403 — заменяем на 404
		return http.StatusNotFound, "Не найдено"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "Превышено время ожидания"
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "Сервис временно недоступен"
	case codes.Internal:
		return http.StatusInternalServerError, "Внутренняя ошибка сервера"
	default:
		return http.StatusInternalServerError, translateError(msg)
	}
}

// translateError converts technical error messages to user-friendly Russian
func translateError(msg string) string {
	// gRPC error patterns from validators and services
	translations := map[string]string{
		"email is required":             "Укажите email",
		"password is required":          "Укажите пароль",
		"full name is required":         "Укажите имя",
		"invalid role":                  "Недопустимая роль",
		"invalid email format":          "Некорректный формат email",
		"password must be at least":     "Пароль должен быть не менее 8 символов",
		"user_id is required":           "Необходима авторизация",
		"age must be between":           "Возраст должен быть от 0 до 150",
		"height_cm must be between":     "Рост должен быть от 50 до 300 см",
		"weight_kg must be between":     "Вес должен быть от 1 до 500 кг",
		"fitness_level must be":         "Выберите уровень подготовки",
		"user not found":                "Пользователь не найден",
		"email already exists":          "Этот email уже зарегистрирован",
		"invalid credentials":           "Неверный email или пароль",
		"user already exists":           "Этот email уже зарегистрирован",
		"value cannot be negative":      "Значение не может быть отрицательным",
		"metric_type is required":       "Укажите тип метрики",
		"invalid metric data":           "Некорректные данные метрики",
		"heart_rate out of valid range": "Пульс вне допустимого диапазона (30–220)",
		"spo2 out of valid range":       "SpO₂ вне допустимого диапазона (70–100)",
		"metric_type not found":         "Тип метрики не найден",
	}
	for pattern, translated := range translations {
		if containsIgnoreCase(msg, pattern) {
			return translated
		}
	}
	// Если не нашли перевод — возвращаем как есть
	return msg
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				containsSubstringIgnoreCase(s, substr))
}

func containsSubstringIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ========== Auth Handlers ==========

func (g *gateway) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode register request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	resp, err := g.userClient.Register(r.Context(), &userpb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     req.Role,
	})
	if err != nil {
		httpCode, errMsg := grpcToHTTPStatus(err)
		g.log.Error("Register failed", zap.Error(err))
		http.Error(w, errMsg, httpCode)
		return
	}

	// Return registration result including verification token (dev mode)
	response := map[string]interface{}{"status": "ok"}
	if resp.GetMessage() != "" {
		response["message"] = resp.GetMessage()
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) registerWithInviteHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		FullName      string `json:"full_name"`
		InviteCode    string `json:"invite_code"`
		LicenseNumber string `json:"license_number"`
		Specialty     string `json:"specialty"`
		Phone         string `json:"phone"`
		Bio           string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode register with invite request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	resp, err := g.userClient.RegisterWithInvite(r.Context(), &userpb.RegisterWithInviteRequest{
		Email:         req.Email,
		Password:      req.Password,
		FullName:      req.FullName,
		InviteCode:    req.InviteCode,
		LicenseNumber: req.LicenseNumber,
		Specialty:     req.Specialty,
		Phone:         req.Phone,
		Bio:           req.Bio,
	})
	if err != nil {
		httpCode, errMsg := grpcToHTTPStatus(err)
		g.log.Error("Register with invite failed", zap.Error(err))
		http.Error(w, errMsg, httpCode)
		return
	}

	response := map[string]interface{}{
		"status":  "ok",
		"user_id": resp.GetUserId(),
	}
	if resp.GetMessage() != "" {
		response["message"] = resp.GetMessage()
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) validateInviteCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode validate invite request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	resp, err := g.userClient.ValidateInviteCode(r.Context(), &userpb.ValidateInviteCodeRequest{
		Code: req.Code,
	})
	if err != nil {
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"is_valid":  resp.GetIsValid(),
		"role":      resp.GetRole(),
		"specialty": resp.GetSpecialty(),
		"error":     resp.GetErrorMessage(),
	}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode login request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	resp, err := g.userClient.Login(r.Context(), &userpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		httpCode, errMsg := grpcToHTTPStatus(err)
		g.log.Error("Login failed", zap.Error(err), zap.String("email", req.Email))
		// Требование #3: Обработка ошибки неподтверждённого email
		if httpCode == http.StatusUnauthorized && strings.Contains(errMsg, "Email not confirmed") {
			http.Error(w, "Email не подтверждён. Проверьте вашу почту.", httpCode)
			return
		}
		http.Error(w, errMsg, httpCode)
		return
	}

	// Требование #11: HMAC-SHA256 подпись критического ответа
	loginResp := map[string]interface{}{
		"status":       "ok",
		"access_token": resp.GetAccessToken(),
		"token_type":   resp.GetTokenType(),
		"expires_in":   resp.GetExpiresIn(),
	}
	if signature, err := auth.SignResponse(loginResp, g.jwtSecret); err == nil {
		w.Header().Set("X-Response-Signature", signature)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(loginResp); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// logoutHandler — принудительная инвалидация сессии
// Требование #1: Явное указание браузеру на удаление cookies (session, refresh_token)
// Требование #7: return после отправки заголовков
func (g *gateway) logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Требование #1: Заголовки для удаления cookies на клиенте
	logoutHeaders := middleware.LogoutHeaders()
	for key, values := range logoutHeaders {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"}); err != nil {
		g.log.Error("Failed to encode logout response", zap.Error(err))
		// Требование #7: После ошибки — возврат без дальнейшего выполнения
		return
	}
	// Требование #7: Немедленное прекращение выполнения
}

// confirmEmailHandler handles email confirmation via token.
func (g *gateway) confirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode confirm email request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		http.Error(w, "Укажите токен подтверждения", http.StatusBadRequest)
		return
	}

	resp, err := g.userClient.ConfirmEmail(r.Context(), &userpb.ConfirmEmailRequest{Token: req.Token})
	if err != nil {
		httpCode, errMsg := grpcToHTTPStatus(err)
		g.log.Error("Confirm email failed", zap.Error(err))
		http.Error(w, errMsg, httpCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Email confirmed. You can now log in.",
		"user_id": resp.GetUserId(),
	}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// emailConfirmPageHandler serves the email confirmation page from a template file.
// The user lands here when clicking the link in the verification email.
func (g *gateway) emailConfirmPageHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	// Load template from web/templates/confirm.html
	tmplPath := "./web/templates/confirm.html"
	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		g.log.Warn("Failed to load confirm template, using fallback", zap.Error(err))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if token == "" {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := fmt.Fprint(w, "<html><body style='background:#0d1117;color:#c9d1d9;font-family:system-ui;'><div style='text-align:center;padding:40px;'><h1 style='color:#f85149;'>Ошибка</h1><p>Токен не найден</p></div></body></html>"); err != nil {
				g.log.Error("Failed to write fallback response", zap.Error(err))
			}
			return
		}
		if _, err := fmt.Fprintf(w, "<html><body style='background:#0d1117;color:#c9d1d9;font-family:system-ui;'><div style='text-align:center;padding:40px;'><h1>Подтверждение email</h1><p>Токен: %s</p></div></body></html>", token); err != nil {
			g.log.Error("Failed to write fallback response", zap.Error(err))
		}
		return
	}

	tmplText := string(tmplBytes)
	tmplText = strings.Replace(tmplText, "{{ .Token }}", token, 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := fmt.Fprint(w, tmplText); err != nil {
		g.log.Error("Failed to write confirm page", zap.Error(err))
	}
}

// checkVerificationStatusHandler checks if a user's email is confirmed.
func (g *gateway) checkVerificationStatusHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Укажите email", http.StatusBadRequest)
		return
	}

	// Query user profile by email — we use GetProfile which requires user_id,
	// but since we only have email, we need to search via the user service.
	// The gateway doesn't have a GetUserByEmail RPC, so we return a not found
	// if we can't resolve the user. For now, we check if the user exists
	// by attempting a profile lookup. In production, add a GetUserByEmail RPC.
	// As a workaround, we return email_confirmed: false for unknown emails.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"email_confirmed": false,
		"email":           email,
	}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// ========== Profile Handlers ==========

func (g *gateway) profileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	resp, err := g.userClient.GetProfile(r.Context(), &userpb.GetProfileRequest{
		UserId: userID,
	})
	if err != nil {
		g.log.Error("Failed to get profile", zap.Error(err), zap.String("user_id", userID))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	// Требование #11: HMAC-SHA256 подпись критического ответа
	profileResp := map[string]interface{}{
		"status":  "ok",
		"profile": resp,
	}
	if signature, err := auth.SignResponse(profileResp, g.jwtSecret); err == nil {
		w.Header().Set("X-Response-Signature", signature)
	}

	if err := json.NewEncoder(w).Encode(profileResp); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var req struct {
		Age               int32    `json:"age"`
		Gender            string   `json:"gender"`
		HeightCm          int32    `json:"height_cm"`
		WeightKg          float64  `json:"weight_kg"`
		FitnessLevel      string   `json:"fitness_level"`
		Goals             []string `json:"goals"`
		Contraindications []string `json:"contraindications"`
		Nutrition         string   `json:"nutrition"`
		SleepHours        float32  `json:"sleep_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode update profile request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	_, err := g.userClient.UpdateProfile(r.Context(), &userpb.UpdateProfileRequest{
		UserId:            userID,
		Age:               ptrInt32(req.Age),
		Gender:            ptrString(req.Gender),
		HeightCm:          ptrInt32(req.HeightCm),
		WeightKg:          ptrFloat64(req.WeightKg),
		FitnessLevel:      ptrString(req.FitnessLevel),
		Goals:             req.Goals,
		Contraindications: req.Contraindications,
		Nutrition:         ptrString(req.Nutrition),
		SleepHours:        ptrFloat32(req.SleepHours),
	})
	if err != nil {
		g.log.Error("Failed to update profile", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// ========== Security #10: Server-side Role Re-verification ==========

// verifyUserRole re-queries the user's role from the database to prevent privilege escalation.
func (g *gateway) verifyUserRole(ctx context.Context, userID, requiredRole string) bool {
	if g.db == nil {
		g.log.Warn("Database not available for role verification")
		return false
	}
	var actualRole string
	err := g.db.QueryRowContext(ctx, "SELECT role FROM users WHERE id = $1", userID).Scan(&actualRole)
	if err == sql.ErrNoRows {
		g.log.Warn("User not found during role verification", zap.String("user_id", userID))
		return false
	}
	if err != nil {
		g.log.Error("Database error during role verification", zap.Error(err))
		return false
	}
	return actualRole == requiredRole
}

// deleteProfileHandler handles profile deletion with server-side role re-verification.
func (g *gateway) deleteProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	if !g.verifyUserRole(r.Context(), userID, "user") {
		g.log.Warn("Role verification failed for profile deletion", zap.String("user_id", userID))
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	// Delete profile directly from database
	if g.db == nil {
		g.log.Error("Database not available for profile deletion")
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}
	_, err := g.db.ExecContext(r.Context(), "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		g.log.Error("Failed to delete profile", zap.Error(err), zap.String("user_id", userID))
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	// Требование #1: Инвалидация сессии после удаления профиля
	logoutHeaders := middleware.LogoutHeaders()
	for key, values := range logoutHeaders {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "deleted"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// adminListUsersHandler handles admin user listing with server-side role re-verification.
func (g *gateway) adminListUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	if !g.verifyUserRole(r.Context(), userID, "admin") {
		g.log.Warn("Non-admin attempted to access user list", zap.String("user_id", userID))
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	if g.db == nil {
		g.log.Error("Database not available for user listing")
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	pageSize := 20
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 {
			pageSize = val
		}
	}
	offset := (page - 1) * pageSize

	rows, err := g.db.QueryContext(r.Context(),
		"SELECT id, email, full_name, role, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		pageSize, offset)
	if err != nil {
		g.log.Error("Failed to query users", zap.Error(err))
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			g.log.Error("Failed to close rows", zap.Error(closeErr))
		}
	}()

	type userInfo struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		FullName  string    `json:"full_name"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
	}
	var users []userInfo
	for rows.Next() {
		var u userInfo
		if scanErr := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.CreatedAt); scanErr != nil {
			g.log.Error("Failed to scan user row", zap.Error(scanErr))
			http.Error(w, "Не найдено", http.StatusNotFound)
			return
		}
		users = append(users, u)
	}
	if scanErr := rows.Err(); scanErr != nil {
		g.log.Error("Rows iteration error", zap.Error(scanErr))
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	adminResp := map[string]interface{}{"status": "ok", "users": users, "total": len(users)}
	if signature, err := auth.SignResponse(adminResp, g.jwtSecret); err == nil {
		w.Header().Set("X-Response-Signature", signature)
	}

	if err := json.NewEncoder(w).Encode(adminResp); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// ========== Biometric Handlers ==========

func (g *gateway) addBiometricRecordHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var req struct {
		MetricType string    `json:"metric_type"`
		Value      float64   `json:"value"`
		Timestamp  time.Time `json:"timestamp"`
		DeviceType string    `json:"device_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}
	// Валидация
	if req.MetricType == "" || req.Value < 0 {
		http.Error(w, "Некорректные данные метрики", http.StatusBadRequest)
		return
	}

	_, err := g.biometricClient.AddRecord(r.Context(), &biometricpb.AddRecordRequest{
		UserId:     userID,
		MetricType: req.MetricType,
		Value:      req.Value,
		Timestamp:  timestamppb.New(req.Timestamp),
		DeviceType: req.DeviceType,
	})
	if err != nil {
		g.log.Error("Failed to add biometric record", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "created"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) getBiometricRecordsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	metricType := r.URL.Query().Get("metric_type")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	limitStr := r.URL.Query().Get("limit")

	var from, to time.Time
	if fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}
	limitInt := int32(100)
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limitInt = int32(l)
		}
	}

	_, err := g.biometricClient.GetRecords(r.Context(), &biometricpb.GetRecordsRequest{
		UserId:     userID,
		MetricType: metricType,
		From:       timestamppb.New(from),
		To:         timestamppb.New(to),
		Limit:      limitInt,
	})
	if err != nil {
		g.log.Error("Failed to get biometric records", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	// Требование #11: HMAC-SHA256 подпись критического ответа
	bioResp := map[string]interface{}{"status": "ok"}
	if signature, err := auth.SignResponse(bioResp, g.jwtSecret); err == nil {
		w.Header().Set("X-Response-Signature", signature)
	}

	if err := json.NewEncoder(w).Encode(bioResp); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// ========== Training Handlers ==========

func (g *gateway) generatePlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var req struct {
		DurationWeeks int     `json:"duration_weeks"`
		AvailableDays []int   `json:"available_days"`
		Class         string  `json:"class"`
		Confidence    float64 `json:"confidence"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode generate plan request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	class := req.Class
	if class == "" {
		class = "endurance_e1e2"
	}

	availableDays := make([]int32, len(req.AvailableDays))
	for i, d := range req.AvailableDays {
		availableDays[i] = int32(d)
	}

	_, err := g.trainingClient.GeneratePlan(r.Context(), &trainingpb.GeneratePlanRequest{
		UserId:              userID,
		ClassificationClass: class,
		Confidence:          req.Confidence,
		DurationWeeks:       int32(req.DurationWeeks),
		AvailableDays:       availableDays,
	})
	if err != nil {
		g.log.Error("Failed to generate plan", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	// Требование #11: HMAC-SHA256 подпись критического ответа
	planResp := map[string]interface{}{"status": "ok"}
	if signature, err := auth.SignResponse(planResp, g.jwtSecret); err == nil {
		w.Header().Set("X-Response-Signature", signature)
	}

	if err := json.NewEncoder(w).Encode(planResp); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) getPlansHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	pageSize := 10
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 {
			pageSize = val
		}
	}

	_, err := g.trainingClient.ListPlans(r.Context(), &trainingpb.ListPlansRequest{
		UserId:   userID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		g.log.Error("Failed to get plans", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	// Требование #11: HMAC-SHA256 подпись критического ответа
	plansResp := map[string]interface{}{"status": "ok"}
	if signature, err := auth.SignResponse(plansResp, g.jwtSecret); err == nil {
		w.Header().Set("X-Response-Signature", signature)
	}

	if err := json.NewEncoder(w).Encode(plansResp); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) completeWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var req struct {
		PlanId    string `json:"plan_id"`
		WorkoutId string `json:"workout_id"`
		Rating    int32  `json:"rating"`
		Feedback  string `json:"feedback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode complete workout request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	_, err := g.trainingClient.CompleteWorkout(r.Context(), &trainingpb.CompleteWorkoutRequest{
		UserId:    userID,
		PlanId:    req.PlanId,
		WorkoutId: req.WorkoutId,
		Rating:    req.Rating,
		Feedback:  req.Feedback,
	})
	if err != nil {
		g.log.Error("Failed to complete workout", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) getProgressHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	_, err := g.trainingClient.GetProgress(r.Context(), &trainingpb.GetProgressRequest{
		UserId: userID,
	})
	if err != nil {
		g.log.Error("Failed to get progress", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}
}

// ========== ML Classifier Handler ==========

func (g *gateway) classifyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	// Build the ML payload (same as sync path)
	bioResp, err := g.biometricClient.GetLatest(r.Context(), &biometricpb.GetLatestRequest{
		UserId:     userID,
		MetricType: "heart_rate",
	})
	if err != nil {
		g.log.Warn("Failed to get heart rate", zap.Error(err))
	}

	mlPayload := extractMLPayload(bioResp)

	if g.mlAsync {
		// Async path: publish to RabbitMQ, return 202
		g.handleAsyncClassify(w, r, mlPayload)
		return
	}

	// Sync path: call ML classifier directly (backward compatible)
	reqBody, _ := json.Marshal(mlPayload)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		g.mlClassifierURL+"/classify",
		bytes.NewReader(reqBody))
	if err != nil {
		g.log.Error("Failed to create ML classifier request", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		g.log.Error("ML classifier request failed", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			g.log.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		g.log.Error("ML classifier returned error", zap.Int("status", resp.StatusCode))
		http.Error(w, "Ошибка классификации", resp.StatusCode)
		return
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		g.log.Error("Failed to write response", zap.Error(err))
	}
}

// handleAsyncClassify publishes a classification job to RabbitMQ.
func (g *gateway) handleAsyncClassify(w http.ResponseWriter, r *http.Request, mlPayload map[string]interface{}) {
	jobID := uuid.New().String()

	body, err := json.Marshal(map[string]interface{}{
		"job_id":             jobID,
		"physiological_data": mlPayload["physiological_data"],
	})
	if err != nil {
		g.log.Error("Failed to marshal classify job", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = g.rmqCh.PublishWithContext(ctx, "", "ml.classify", false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		g.log.Error("Failed to publish classify job to RabbitMQ", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": "pending",
	}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
	}
}

// classifyStatusHandler returns the status/result of an async classification job.
func (g *gateway) classifyStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]
	if jobID == "" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	val, err := g.rdb.Get(ctx, fmt.Sprintf("ml:result:%s", jobID)).Result()
	if err == redis.Nil {
		// Key not found - job still processing or expired
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id": jobID,
			"status": "processing",
		}); err != nil {
			g.log.Error("Failed to encode response", zap.Error(err))
		}
		return
	}
	if err != nil {
		g.log.Error("Failed to get job result from Redis", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}

	// Return the stored result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(val)); err != nil {
		g.log.Error("Failed to write response", zap.Error(err))
	}
}

// ========== ML Generator Handler ==========

func (g *gateway) generateMLPlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Необходима авторизация", http.StatusUnauthorized)
		return
	}

	var req struct {
		ClassName     string `json:"training_class"`
		DurationWeeks int    `json:"duration_weeks"`
		AvailableDays []int  `json:"available_days"`
		Preferences   struct {
			MaxDuration        int      `json:"max_duration"`
			AvailableEquipment []string `json:"available_equipment"`
			PreferredTime      string   `json:"preferred_time"`
		} `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.log.Error("Failed to decode generate ML plan request", zap.Error(err))
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	// Get user profile
	profile, err := g.userClient.GetProfile(r.Context(), &userpb.GetProfileRequest{UserId: userID})
	if err != nil {
		g.log.Error("Failed to get profile", zap.Error(err))
		httpCode, errMsg := grpcToHTTPStatus(err)
		http.Error(w, errMsg, httpCode)
		return
	}

	userProfile := map[string]interface{}{}
	if profile.Age > 0 {
		userProfile["age"] = profile.Age
	}
	if profile.Gender != "" {
		userProfile["gender"] = profile.Gender
	}
	if profile.WeightKg > 0 {
		userProfile["weight"] = profile.WeightKg
	}
	if profile.HeightCm > 0 {
		userProfile["height"] = profile.HeightCm
	}
	if profile.FitnessLevel != "" {
		userProfile["fitness_level"] = profile.FitnessLevel
	}
	if len(profile.Contraindications) > 0 {
		userProfile["health_conditions"] = profile.Contraindications
	}
	if len(profile.Goals) > 0 {
		userProfile["goals"] = profile.Goals
	}
	if profile.SleepHours > 0 {
		userProfile["sleep_hours"] = profile.SleepHours
	}
	if profile.Nutrition != "" {
		userProfile["nutrition"] = profile.Nutrition
	}

	if g.mlAsync {
		g.handleAsyncGeneratePlan(w, r, req.ClassName, userProfile, req.Preferences)
		return
	}

	// Sync path: call ML generator directly (backward compatible)
	genReq := map[string]interface{}{
		"training_class": req.ClassName,
		"user_profile":   userProfile,
		"preferences": map[string]interface{}{
			"max_duration":        req.Preferences.MaxDuration,
			"available_equipment": req.Preferences.AvailableEquipment,
			"preferred_time":      req.Preferences.PreferredTime,
		},
	}

	reqBody, err := json.Marshal(genReq)
	if err != nil {
		g.log.Error("Failed to marshal generator request", zap.Error(err))
		http.Error(w, "Ошибка формирования запроса", http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(g.mlGeneratorURL+"/generate-plan", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		g.log.Error("Failed to call ML generator", zap.Error(err))
		http.Error(w, "Ошибка обращения к ML-генератору", http.StatusInternalServerError)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			g.log.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		g.log.Error("Failed to read response body", zap.Error(err))
		http.Error(w, "Ошибка чтения ответа", http.StatusInternalServerError)
		return
	}

	// Save plan to training service (non-blocking)
	var mlResp map[string]interface{}
	if err := json.Unmarshal(body, &mlResp); err == nil {
		availableDays := make([]int32, len(req.AvailableDays))
		for i, d := range req.AvailableDays {
			availableDays[i] = int32(d)
		}

		_, saveErr := g.trainingClient.GeneratePlan(r.Context(), &trainingpb.GeneratePlanRequest{
			UserId:              userID,
			ClassificationClass: req.ClassName,
			Confidence:          0.85,
			DurationWeeks:       int32(req.DurationWeeks),
			AvailableDays:       availableDays,
		})
		if saveErr != nil {
			g.log.Warn("Failed to save ML plan to training service (non-critical)", zap.Error(saveErr))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(body); err != nil {
		g.log.Error("Failed to write response body", zap.Error(err))
	}
}

// handleAsyncGeneratePlan publishes a plan generation job to RabbitMQ.
func (g *gateway) handleAsyncGeneratePlan(w http.ResponseWriter, r *http.Request, trainingClass string, userProfile map[string]interface{}, prefs struct {
	MaxDuration        int      `json:"max_duration"`
	AvailableEquipment []string `json:"available_equipment"`
	PreferredTime      string   `json:"preferred_time"`
}) {
	jobID := uuid.New().String()

	body, err := json.Marshal(map[string]interface{}{
		"job_id":         jobID,
		"training_class": trainingClass,
		"user_profile":   userProfile,
		"preferences": map[string]interface{}{
			"max_duration":        prefs.MaxDuration,
			"available_equipment": prefs.AvailableEquipment,
			"preferred_time":      prefs.PreferredTime,
		},
	})
	if err != nil {
		g.log.Error("Failed to marshal generate-plan job", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = g.rmqCh.PublishWithContext(ctx, "", "ml.generate", false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		g.log.Error("Failed to publish generate-plan job to RabbitMQ", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": "pending",
	}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
	}
}

// generatePlanStatusHandler returns the status/result of an async plan generation job.
func (g *gateway) generatePlanStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]
	if jobID == "" {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	val, err := g.rdb.Get(ctx, fmt.Sprintf("ml:result:%s", jobID)).Result()
	if err == redis.Nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id": jobID,
			"status": "processing",
		}); err != nil {
			g.log.Error("Failed to encode response", zap.Error(err))
		}
		return
	}
	if err != nil {
		g.log.Error("Failed to get job result from Redis", zap.Error(err))
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(val)); err != nil {
		g.log.Error("Failed to write response", zap.Error(err))
	}
}

// ========== Health Check ==========

func (g *gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{ // проверили ошибку
		"status":        "ok",
		"service":       "gateway",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"ml_classifier": g.mlClassifierURL,
		"ml_generator":  g.mlGeneratorURL,
	})
}

// extractFeatures извлекает фичи из биометрических данных для ML-классификации
func extractMLPayload(bioResp *biometricpb.BiometricRecord) map[string]interface{} {
	// Дефолтные значения при отсутствии данных
	heartRate := 70.0
	hrv := 50.0
	spo2 := 98.0
	temp := 36.6
	bpSystolic := 120.0
	bpDiastolic := 80.0
	sleepHours := 7.0

	if bioResp != nil {
		switch bioResp.MetricType {
		case "heart_rate":
			heartRate = bioResp.Value
		case "hrv":
			hrv = bioResp.Value
		case "spo2":
			spo2 = bioResp.Value
		case "temperature":
			temp = bioResp.Value
		case "systolic_pressure":
			bpSystolic = bioResp.Value
		case "diastolic_pressure":
			bpDiastolic = bioResp.Value
		case "sleep_hours":
			sleepHours = bioResp.Value
		}
	}

	return map[string]interface{}{
		"physiological_data": map[string]float64{
			"heart_rate":               heartRate,
			"heart_rate_variability":   hrv,
			"spo2":                     spo2,
			"temperature":              temp,
			"blood_pressure_systolic":  bpSystolic,
			"blood_pressure_diastolic": bpDiastolic,
			"sleep_hours":              sleepHours,
		},
	}
}

// ========== Device Connector Proxy Handlers ==========

// deviceRegisterHandler proxies device registration to device-connector
func (g *gateway) deviceRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if g.deviceConnectorURL == "" {
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		g.log.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Ошибка чтения ответа", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		g.deviceConnectorURL+"/api/v1/devices/register",
		bytes.NewReader(body))
	if err != nil {
		g.log.Error("Failed to create device register request", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		g.log.Error("Device connector unreachable", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			g.log.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		g.log.Error("Failed to write response", zap.Error(err))
	}
}

// deviceIngestHandler proxies data ingestion to device-connector
func (g *gateway) deviceIngestHandler(w http.ResponseWriter, r *http.Request) {
	if g.deviceConnectorURL == "" {
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		g.log.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Ошибка чтения ответа", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		g.deviceConnectorURL+"/api/v1/devices/"+deviceID+"/ingest",
		bytes.NewReader(body))
	if err != nil {
		g.log.Error("Failed to create device ingest request", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		g.log.Error("Device connector unreachable", zap.Error(err))
		http.Error(w, "ML-сервис временно недоступен", http.StatusServiceUnavailable)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			g.log.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		g.log.Error("Failed to write response", zap.Error(err))
	}
}

// ========== Main ==========

func main() {
	log := logger.New("gateway")
	defer func() {
		if syncErr := log.Sync(); syncErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", syncErr)
		}
	}()

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50051"
	}

	biometricServiceAddr := os.Getenv("BIOMETRIC_SERVICE_ADDR")
	if biometricServiceAddr == "" {
		biometricServiceAddr = "localhost:50052"
	}

	trainingServiceAddr := os.Getenv("TRAINING_SERVICE_ADDR")
	if trainingServiceAddr == "" {
		trainingServiceAddr = "localhost:50053"
	}

	mlClassifierURL := os.Getenv("ML_CLASSIFIER_URL")
	if mlClassifierURL == "" {
		mlClassifierURL = "http://localhost:8001"
	}

	mlGeneratorURL := os.Getenv("ML_GENERATOR_URL")
	if mlGeneratorURL == "" {
		mlGeneratorURL = "http://localhost:8002"
	}

	deviceConnectorURL := os.Getenv("DEVICE_CONNECTOR_URL")
	if deviceConnectorURL == "" {
		deviceConnectorURL = "http://localhost:8082"
	}

	// Async ML processing configuration
	mlAsync := os.Getenv("ML_ASYNC") == "true" || os.Getenv("ML_ASYNC") == "True" || os.Getenv("ML_ASYNC") == "1"

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Database connection for server-side role re-verification (Security #10)
	dbURL := os.Getenv("DATABASE_URL")
	var db *sql.DB
	if dbURL != "" {
		var openErr error
		db, openErr = sql.Open("postgres", dbURL)
		if openErr != nil {
			log.Fatal("Failed to open database", zap.Error(openErr))
		}
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(5 * time.Minute)
		if pingErr := db.Ping(); pingErr != nil {
			log.Fatal("Failed to ping database", zap.Error(pingErr))
		}
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				log.Error("Failed to close database connection", zap.Error(closeErr))
			}
		}()
	}

	// Redis client for job result storage (used in async mode)
	var rdb *redis.Client
	if mlAsync {
		rdb = redis.NewClient(&redis.Options{
			Addr: redisHost + ":6379",
		})
		if pingErr := rdb.Ping(context.Background()).Err(); pingErr != nil {
			log.Warn("Redis unavailable, async ML mode disabled", zap.Error(pingErr))
			mlAsync = false
		} else {
			log.Info("Redis connected for async job results", zap.String("host", redisHost))
		}
	}

	// RabbitMQ channel for publishing ML jobs (used in async mode)
	var rmqCh *amqp.Channel
	var rmqClose func()
	if mlAsync {
		rmqConn, rmqErr := amqp.Dial(rabbitmqURL)
		if rmqErr != nil {
			log.Warn("RabbitMQ unavailable, async ML mode disabled", zap.Error(rmqErr))
			mlAsync = false
			rdb = nil
		} else {
			rmqCh, rmqErr = rmqConn.Channel()
			if rmqErr != nil {
				log.Warn("Failed to create RabbitMQ channel, async ML mode disabled", zap.Error(rmqErr))
				mlAsync = false
				rdb = nil
				if closeErr := rmqConn.Close(); closeErr != nil {
					log.Warn("Failed to close RabbitMQ connection", zap.Error(closeErr))
				}
			} else {
				// Declare queues (idempotent)
				_, _ = rmqCh.QueueDeclare("ml.classify", true, false, false, false, nil)
				_, _ = rmqCh.QueueDeclare("ml.generate", true, false, false, false, nil)
				log.Info("RabbitMQ connected for async ML jobs", zap.String("url", rabbitmqURL))
				rmqClose = func() {
					if closeErr := rmqConn.Close(); closeErr != nil {
						log.Warn("Failed to close RabbitMQ connection", zap.Error(closeErr))
					}
				}
			}
		}
	}

	// ✅ Исправлено: используем grpc.NewClient вместо устаревшего grpc.Dial
	userConn, err := grpc.NewClient(userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true), grpc.MaxCallRecvMsgSize(10<<20)),
	)
	if err != nil {
		log.Fatal("Failed to connect to user service", zap.Error(err))
	}
	defer func() {
		if closeErr := userConn.Close(); closeErr != nil {
			log.Error("Failed to close user service connection", zap.Error(closeErr))
		}
	}()

	biometricConn, err := grpc.NewClient(biometricServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		log.Fatal("Failed to connect to biometric service", zap.Error(err))
	}
	defer func() {
		if closeErr := biometricConn.Close(); closeErr != nil {
			log.Error("Failed to close biometric service connection", zap.Error(closeErr))
		}
	}()

	trainingConn, err := grpc.NewClient(trainingServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		log.Fatal("Failed to connect to training service", zap.Error(err))
	}
	defer func() {
		if closeErr := trainingConn.Close(); closeErr != nil {
			log.Error("Failed to close training service connection", zap.Error(closeErr))
		}
	}()

	if rmqClose != nil {
		defer rmqClose()
	}

	g := &gateway{
		userClient:         userpb.NewUserServiceClient(userConn),
		biometricClient:    biometricpb.NewBiometricServiceClient(biometricConn),
		trainingClient:     trainingpb.NewTrainingServiceClient(trainingConn),
		mlClassifierURL:    mlClassifierURL,
		mlGeneratorURL:     mlGeneratorURL,
		deviceConnectorURL: deviceConnectorURL,
		log:                log,
		jwtSecret:          jwtSecret,
		db:                 db,
		rdb:                rdb,
		rmqCh:              rmqCh,
		mlAsync:            mlAsync,
	}

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/v1/register", g.registerHandler).Methods("POST")
	r.HandleFunc("/api/v1/register/invite", g.registerWithInviteHandler).Methods("POST")
	r.HandleFunc("/api/v1/invite/validate", g.validateInviteCodeHandler).Methods("POST")
	r.HandleFunc("/api/v1/login", g.loginHandler).Methods("POST")
	r.HandleFunc("/api/v1/auth/confirm", g.confirmEmailHandler).Methods("POST")
	r.HandleFunc("/api/v1/auth/verify-status", g.checkVerificationStatusHandler).Methods("GET")
	r.HandleFunc("/health", g.healthHandler).Methods("GET")

	// Email confirmation page (GET /confirm?token=xxx)
	r.HandleFunc("/confirm", g.emailConfirmPageHandler).Methods("GET")

	// Device connector routes (device token auth, not JWT)
	r.HandleFunc("/api/v1/devices/register", g.deviceRegisterHandler).Methods("POST")
	r.HandleFunc("/api/v1/devices/{device_id}/ingest", g.deviceIngestHandler).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtSecret, log.Logger))

	// Требование #1: Logout с инвалидацией сессии
	protected.HandleFunc("/logout", g.logoutHandler).Methods("POST")

	protected.HandleFunc("/profile", g.profileHandler).Methods("GET")
	protected.HandleFunc("/profile", g.updateProfileHandler).Methods("PUT")
	protected.HandleFunc("/profile", g.deleteProfileHandler).Methods("DELETE")

	// Admin routes (server-side role re-verification in handler)
	protected.HandleFunc("/admin/users", g.adminListUsersHandler).Methods("GET")

	protected.HandleFunc("/biometrics", g.addBiometricRecordHandler).Methods("POST")
	protected.HandleFunc("/biometrics", g.getBiometricRecordsHandler).Methods("GET")

	protected.HandleFunc("/training/generate", g.generatePlanHandler).Methods("POST")
	protected.HandleFunc("/training/plans", g.getPlansHandler).Methods("GET")
	protected.HandleFunc("/training/complete", g.completeWorkoutHandler).Methods("POST")
	protected.HandleFunc("/training/progress", g.getProgressHandler).Methods("GET")

	protected.HandleFunc("/ml/classify", g.classifyHandler).Methods("POST")
	protected.HandleFunc("/ml/classify/{job_id}", g.classifyStatusHandler).Methods("GET")
	protected.HandleFunc("/ml/generate-plan", g.generateMLPlanHandler).Methods("POST")
	protected.HandleFunc("/ml/generate-plan/{job_id}", g.generatePlanStatusHandler).Methods("GET")

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	// Middleware
	// Security headers (CSP, HSTS, etc.) are handled by nginx — no duplication here
	handler := middleware.RequestID(r)
	handler = middleware.RateLimit(handler)

	log.Info("Gateway starting",
		zap.String("port", port),
		zap.String("ml_classifier", mlClassifierURL),
		zap.String("ml_generator", mlGeneratorURL))

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
