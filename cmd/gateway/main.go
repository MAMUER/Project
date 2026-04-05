package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	biometricpb "github.com/MAMUER/Project/api/gen/biometric"
	trainingpb "github.com/MAMUER/Project/api/gen/training"
	userpb "github.com/MAMUER/Project/api/gen/user"
	"github.com/MAMUER/Project/internal/auth"
	"github.com/MAMUER/Project/internal/logger"
	"github.com/MAMUER/Project/internal/middleware"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type gateway struct {
	userClient      userpb.UserServiceClient
	biometricClient biometricpb.BiometricServiceClient
	trainingClient  trainingpb.TrainingServiceClient
	mlClassifierURL string
	mlGeneratorURL  string
	log             *logger.Logger
	jwtSecret       string
}

// ========== Helper Functions ==========

func ptrInt32(v int32) *int32       { return &v }
func ptrString(v string) *string    { return &v }
func ptrFloat64(v float64) *float64 { return &v }
func ptrFloat32(v float32) *float32 { return &v }

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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := g.userClient.Register(r.Context(), &userpb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     req.Role,
	})
	if err != nil {
		g.log.Error("Register failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err := g.userClient.Login(r.Context(), &userpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		g.log.Error("Login failed", zap.Error(err), zap.String("email", req.Email))
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ========== Profile Handlers ==========

func (g *gateway) profileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := g.userClient.GetProfile(r.Context(), &userpb.GetProfileRequest{
		UserId: userID,
	})
	if err != nil {
		g.log.Error("Failed to get profile", zap.Error(err), zap.String("user_id", userID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signature, err := auth.SignResponse(resp, g.jwtSecret)
	if err == nil {
		w.Header().Set("X-Response-Signature", signature)
	}

	// ✅ Проверяем ошибку кодирования
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "invalid request", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ========== Biometric Handlers ==========

func (g *gateway) addBiometricRecordHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		MetricType string    `json:"metric_type"`
		Value      float64   `json:"value"`
		Timestamp  time.Time `json:"timestamp"`
		DeviceType string    `json:"device_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// Валидация
	if req.MetricType == "" || req.Value < 0 {
		http.Error(w, "invalid metric data", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) getBiometricRecordsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ========== Training Handlers ==========

func (g *gateway) generatePlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "invalid request", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) getPlansHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) completeWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "invalid request", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (g *gateway) getProgressHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := g.trainingClient.GetProgress(r.Context(), &trainingpb.GetProgressRequest{
		UserId: userID,
	})
	if err != nil {
		g.log.Error("Failed to get progress", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Исправлено: проверяем ошибку Encode
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
		g.log.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ========== ML Classifier Handler ==========

func (g *gateway) classifyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем последние биометрические данные
	bioResp, err := g.biometricClient.GetLatest(r.Context(), &biometricpb.GetLatestRequest{
		UserId:     userID,
		MetricType: "heart_rate",
	})
	if err != nil {
		g.log.Warn("Failed to get heart rate", zap.Error(err))
		// Используем дефолтное значение
	}

	// Формируем фичи для классификации
	features := extractFeatures(bioResp)

	// ✅ Создаём контекст с таймаутом
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// ✅ Создаём запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "POST",
		g.mlClassifierURL+"/classify",
		bytes.NewReader(features))
	if err != nil {
		g.log.Error("Failed to create ML classifier request", zap.Error(err))
		http.Error(w, "classification service unavailable", http.StatusServiceUnavailable)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// ✅ Выполняем запрос через http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		g.log.Error("ML classifier request failed", zap.Error(err))
		http.Error(w, "classification service unavailable", http.StatusServiceUnavailable)
		return
	}
	// ✅ Проверяем ошибку закрытия тела ответа
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			g.log.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		g.log.Error("ML classifier returned error", zap.Int("status", resp.StatusCode))
		http.Error(w, "classification failed", resp.StatusCode)
		return
	}

	// ✅ Проверяем ошибку копирования тела
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		g.log.Error("Failed to write response", zap.Error(err))
		// Тело уже частично отправлено, логируем ошибку
	}
}

// ========== ML Generator Handler ==========

func (g *gateway) generateMLPlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Получаем профиль пользователя
	profile, err := g.userClient.GetProfile(r.Context(), &userpb.GetProfileRequest{UserId: userID})
	if err != nil {
		g.log.Error("Failed to get profile", zap.Error(err))
		http.Error(w, "failed to get profile", http.StatusInternalServerError)
		return
	}

	// Формируем запрос к ML генератору
	genReq := map[string]interface{}{
		"training_class": req.ClassName,
		"user_profile": map[string]interface{}{
			"gender":            profile.Gender,
			"age":               profile.Age,
			"fitness_level":     profile.FitnessLevel,
			"weight":            profile.WeightKg,
			"height":            profile.HeightCm,
			"health_conditions": profile.Contraindications,
			"goals":             profile.Goals,
			"sleep_hours":       profile.SleepHours,
			"nutrition":         profile.Nutrition,
			"lifestyle": map[string]interface{}{
				"sleep_hours":       profile.SleepHours,
				"nutrition_quality": 0.7,
			},
		},
		"preferences": map[string]interface{}{
			"max_duration":        req.Preferences.MaxDuration,
			"available_equipment": req.Preferences.AvailableEquipment,
			"preferred_time":      req.Preferences.PreferredTime,
		},
	}

	reqBody, err := json.Marshal(genReq)
	if err != nil {
		g.log.Error("Failed to marshal generator request", zap.Error(err))
		http.Error(w, "failed to prepare request", http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(g.mlGeneratorURL+"/generate-plan", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		g.log.Error("Failed to call ML generator", zap.Error(err))
		http.Error(w, "failed to generate plan", http.StatusInternalServerError)
		return
	}
	// ✅ Проверяем ошибку закрытия тела ответа
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			g.log.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		g.log.Error("Failed to read response body", zap.Error(err))
		http.Error(w, "failed to read response", http.StatusInternalServerError)
		return
	}

	// Сохраняем план в training service
	var mlResp map[string]interface{}
	if err := json.Unmarshal(body, &mlResp); err == nil {
		availableDays := make([]int32, len(req.AvailableDays))
		for i, d := range req.AvailableDays {
			availableDays[i] = int32(d)
		}

		_, err = g.trainingClient.GeneratePlan(r.Context(), &trainingpb.GeneratePlanRequest{
			UserId:              userID,
			ClassificationClass: req.ClassName,
			Confidence:          0.85,
			DurationWeeks:       int32(req.DurationWeeks),
			AvailableDays:       availableDays,
		})
		if err != nil {
			g.log.Warn("Failed to save plan to training service", zap.Error(err))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	// ✅ Проверяем ошибку Write
	if _, err := w.Write(body); err != nil {
		g.log.Error("Failed to write response body", zap.Error(err))
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
func extractFeatures(bioResp *biometricpb.BiometricRecord) []byte { // ← pb → biometricpb
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

	// Формируем JSON с фичами
	features := map[string]float64{
		"heart_rate":               heartRate,
		"heart_rate_variability":   hrv,
		"spo2":                     spo2,
		"temperature":              temp,
		"blood_pressure_systolic":  bpSystolic,
		"blood_pressure_diastolic": bpDiastolic,
		"sleep_hours":              sleepHours,
	}

	result, _ := json.Marshal(features)
	return result
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
		log.Warn("Using default JWT secret")
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

	g := &gateway{
		userClient:      userpb.NewUserServiceClient(userConn),
		biometricClient: biometricpb.NewBiometricServiceClient(biometricConn),
		trainingClient:  trainingpb.NewTrainingServiceClient(trainingConn),
		mlClassifierURL: mlClassifierURL,
		mlGeneratorURL:  mlGeneratorURL,
		log:             log,
		jwtSecret:       jwtSecret,
	}

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/v1/register", g.registerHandler).Methods("POST")
	r.HandleFunc("/api/v1/login", g.loginHandler).Methods("POST")
	r.HandleFunc("/health", g.healthHandler).Methods("GET")

	// Protected routes
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtSecret, log.Logger))

	protected.HandleFunc("/profile", g.profileHandler).Methods("GET")
	protected.HandleFunc("/profile", g.updateProfileHandler).Methods("PUT")

	protected.HandleFunc("/biometrics", g.addBiometricRecordHandler).Methods("POST")
	protected.HandleFunc("/biometrics", g.getBiometricRecordsHandler).Methods("GET")

	protected.HandleFunc("/training/generate", g.generatePlanHandler).Methods("POST")
	protected.HandleFunc("/training/plans", g.getPlansHandler).Methods("GET")
	protected.HandleFunc("/training/complete", g.completeWorkoutHandler).Methods("POST")
	protected.HandleFunc("/training/progress", g.getProgressHandler).Methods("GET")

	protected.HandleFunc("/ml/classify", g.classifyHandler).Methods("POST")
	protected.HandleFunc("/ml/generate-plan", g.generateMLPlanHandler).Methods("POST")

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	// Middleware
	handler := middleware.RequestID(r)
	handler = middleware.RateLimit(handler)
	handler = middleware.RemoveServerHeader(handler)
	handler = middleware.SecurityHeaders(handler)

	log.Info("Gateway starting",
		zap.String("port", port),
		zap.String("ml_classifier", mlClassifierURL),
		zap.String("ml_generator", mlGeneratorURL))

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
