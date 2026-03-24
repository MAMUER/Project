package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	biometricpb "github.com/MAMUER/Project/api/gen/biometric"
	trainingpb "github.com/MAMUER/Project/api/gen/training"
	userpb "github.com/MAMUER/Project/api/gen/user"
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

	resp, err := g.userClient.Register(r.Context(), &userpb.RegisterRequest{
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

	json.NewEncoder(w).Encode(resp)
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

	resp, err := g.userClient.Login(r.Context(), &userpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		g.log.Error("Login failed", zap.Error(err), zap.String("email", req.Email))
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

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

	json.NewEncoder(w).Encode(resp)
}

func (g *gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"service":   "gateway",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

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
		g.log.Error("Failed to decode biometric record request", zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp, err := g.biometricClient.AddRecord(r.Context(), &biometricpb.AddRecordRequest{
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

	json.NewEncoder(w).Encode(resp)
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

	resp, err := g.biometricClient.GetRecords(r.Context(), &biometricpb.GetRecordsRequest{
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

	json.NewEncoder(w).Encode(resp)
}

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
		class = "general"
	}

	availableDays := make([]int32, len(req.AvailableDays))
	for i, d := range req.AvailableDays {
		availableDays[i] = int32(d)
		_ = i // используем i чтобы избежать ошибки unused
		_ = d // используем d чтобы избежать ошибки unused
	}

	resp, err := g.trainingClient.GeneratePlan(r.Context(), &trainingpb.GeneratePlanRequest{
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

	json.NewEncoder(w).Encode(resp)
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

	resp, err := g.trainingClient.ListPlans(r.Context(), &trainingpb.ListPlansRequest{
		UserId:   userID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		g.log.Error("Failed to get plans", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
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

	resp, err := g.trainingClient.CompleteWorkout(r.Context(), &trainingpb.CompleteWorkoutRequest{
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

	json.NewEncoder(w).Encode(resp)
}

func (g *gateway) getProgressHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := g.trainingClient.GetProgress(r.Context(), &trainingpb.GetProgressRequest{
		UserId: userID,
	})
	if err != nil {
		g.log.Error("Failed to get progress", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

// Хендлеры для ML
func (g *gateway) classifyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	records, err := g.biometricClient.GetRecords(r.Context(), &biometricpb.GetRecordsRequest{
		UserId: userID,
		Limit:  100,
	})
	if err != nil {
		http.Error(w, "failed to get biometric data", http.StatusInternalServerError)
		return
	}

	features := map[string]float64{
		"heart_rate":  0,
		"spo2":        0,
		"temperature": 0,
		"sleep_hours": 0,
	}

	var hrCount, spo2Count, tempCount int
	for _, rec := range records.Records {
		switch rec.MetricType {
		case "heart_rate":
			features["heart_rate"] += rec.Value
			hrCount++
		case "spo2":
			features["spo2"] += rec.Value
			spo2Count++
		case "temperature":
			features["temperature"] += rec.Value
			tempCount++
		case "sleep":
			features["sleep_hours"] += rec.Value
		}
	}

	if hrCount > 0 {
		features["heart_rate"] /= float64(hrCount)
	}
	if spo2Count > 0 {
		features["spo2"] /= float64(spo2Count)
	}
	if tempCount > 0 {
		features["temperature"] /= float64(tempCount)
	}

	profile, err := g.userClient.GetProfile(r.Context(), &userpb.GetProfileRequest{UserId: userID})
	if err != nil {
		http.Error(w, "failed to get profile", http.StatusInternalServerError)
		return
	}

	classifyReq := map[string]interface{}{
		"features": map[string]interface{}{
			"heart_rate":               features["heart_rate"],
			"ecg":                      0.8,
			"blood_pressure_systolic":  120,
			"blood_pressure_diastolic": 80,
			"spo2":                     features["spo2"],
			"temperature":              features["temperature"],
			"sleep_hours":              features["sleep_hours"],
		},
		"user_context": map[string]interface{}{
			"age":               profile.Age,
			"gender":            profile.Gender,
			"fitness_level":     profile.FitnessLevel,
			"goals":             profile.Goals,
			"contraindications": profile.Contraindications,
		},
	}

	reqBody, _ := json.Marshal(classifyReq)
	resp, err := http.Post(g.mlClassifierURL+"/classify", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		g.log.Error("Failed to call ML classifier", zap.Error(err))
		http.Error(w, "failed to classify", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (g *gateway) generateMLPlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ClassName     string  `json:"class_name"`
		Confidence    float64 `json:"confidence"`
		DurationWeeks int     `json:"duration_weeks"`
		AvailableDays []int   `json:"available_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	profile, err := g.userClient.GetProfile(r.Context(), &userpb.GetProfileRequest{UserId: userID})
	if err != nil {
		http.Error(w, "failed to get profile", http.StatusInternalServerError)
		return
	}

	genReq := map[string]interface{}{
		"class_name":        req.ClassName,
		"confidence":        req.Confidence,
		"duration_weeks":    req.DurationWeeks,
		"available_days":    req.AvailableDays,
		"user_goals":        profile.Goals,
		"fitness_level":     profile.FitnessLevel,
		"contraindications": profile.Contraindications,
	}

	reqBody, _ := json.Marshal(genReq)
	resp, err := http.Post(g.mlGeneratorURL+"/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		g.log.Error("Failed to call ML generator", zap.Error(err))
		http.Error(w, "failed to generate plan", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var mlResp struct {
		PlanID   string                 `json:"plan_id"`
		PlanData map[string]interface{} `json:"plan_data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &mlResp); err != nil {
		http.Error(w, "failed to parse ML response", http.StatusInternalServerError)
		return
	}

	availableDays := make([]int32, len(req.AvailableDays))
	for i, d := range req.AvailableDays {
		availableDays[i] = int32(d)
		_ = i
		_ = d
	}

	_, err = g.trainingClient.GeneratePlan(r.Context(), &trainingpb.GeneratePlanRequest{
		UserId:              userID,
		ClassificationClass: req.ClassName,
		Confidence:          req.Confidence,
		DurationWeeks:       int32(req.DurationWeeks),
		AvailableDays:       availableDays,
	})
	if err != nil {
		g.log.Error("Failed to save plan to training service", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func main() {
	log := logger.New("gateway")
	defer log.Sync()

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

	userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to user service", zap.Error(err))
	}
	defer userConn.Close()

	biometricConn, err := grpc.Dial(biometricServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to biometric service", zap.Error(err))
	}
	defer biometricConn.Close()

	trainingConn, err := grpc.Dial(trainingServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to training service", zap.Error(err))
	}
	defer trainingConn.Close()

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

	r.HandleFunc("/api/v1/register", g.registerHandler).Methods("POST")
	r.HandleFunc("/api/v1/login", g.loginHandler).Methods("POST")
	r.HandleFunc("/health", g.healthHandler).Methods("GET")

	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtSecret, log.Logger))
	protected.HandleFunc("/profile", g.profileHandler).Methods("GET")

	protected.HandleFunc("/biometrics", g.addBiometricRecordHandler).Methods("POST")
	protected.HandleFunc("/biometrics", g.getBiometricRecordsHandler).Methods("GET")

	protected.HandleFunc("/training/generate", g.generatePlanHandler).Methods("POST")
	protected.HandleFunc("/training/plans", g.getPlansHandler).Methods("GET")
	protected.HandleFunc("/training/complete", g.completeWorkoutHandler).Methods("POST")
	protected.HandleFunc("/training/progress", g.getProgressHandler).Methods("GET")

	protected.HandleFunc("/ml/classify", g.classifyHandler).Methods("GET")
	protected.HandleFunc("/ml/generate-plan", g.generateMLPlanHandler).Methods("POST")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	handler := middleware.RequestID(r)

	log.Info("Gateway starting", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}