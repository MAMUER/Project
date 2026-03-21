package apitypes

import "time"

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token     string `json:"token"`
    UserID    string `json:"user_id"`
    Role      string `json:"role"`
    ExpiresIn int64  `json:"expires_in"`
}

type BiometricRequest struct {
    UserID      string    `json:"user_id"`
    DeviceType  string    `json:"device_type"`
    HeartRate   int       `json:"heart_rate"`
    ECG         string    `json:"ecg"`
    BloodPressure struct {
        Systolic  int `json:"systolic"`
        Diastolic int `json:"diastolic"`
    } `json:"blood_pressure"`
    SpO2        int       `json:"spo2"`
    Temperature float64   `json:"temperature"`
    Sleep       struct {
        Duration  int `json:"duration"`
        DeepSleep int `json:"deep_sleep"`
    } `json:"sleep"`
    Timestamp   time.Time `json:"timestamp"`
}

type GenerateProgramRequest struct {
    TrainingClass      string   `json:"training_class"`
    Contraindications  []string `json:"contraindications"`
    Goals              []string `json:"goals"`
    FitnessLevel       string   `json:"fitness_level"`
    AgeGroup           string   `json:"age_group"`
    Gender             string   `json:"gender"`
    HasInjury          bool     `json:"has_injury"`
    DurationWeeks      int      `json:"duration_weeks"`
}

type TrainingProgram struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Weeks     []Week    `json:"weeks"`
    StartDate string    `json:"start_date"`
    EndDate   string    `json:"end_date"`
    CreatedAt time.Time `json:"created_at"`
}

type Week struct {
    WeekNumber int       `json:"week_number"`
    Schedule   []Workout `json:"schedule"`
    Notes      []string  `json:"notes"`
}

type Workout struct {
    Day             int    `json:"day"`
    WorkoutType     string `json:"workout_type"`
    Intensity       string `json:"intensity"`
    DurationMinutes int    `json:"duration_minutes"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    int    `json:"code"`
    Details string `json:"details,omitempty"`
}

type HealthResponse struct {
    Status  string            `json:"status"`
    Service string            `json:"service"`
    Version string            `json:"version"`
    Uptime  string            `json:"uptime"`
    Checks  map[string]string `json:"checks,omitempty"`
}