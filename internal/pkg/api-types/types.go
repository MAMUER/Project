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
	UserID        string `json:"user_id"`
	DeviceType    string `json:"device_type"`
	HeartRate     int    `json:"heart_rate"`
	ECG           string `json:"ecg"`
	BloodPressure struct {
		Systolic  int `json:"systolic"`
		Diastolic int `json:"diastolic"`
	} `json:"blood_pressure"`
	SpO2        int     `json:"spo2"`
	Temperature float64 `json:"temperature"`
	Sleep       struct {
		Duration  int `json:"duration"`
		DeepSleep int `json:"deep_sleep"`
	} `json:"sleep"`
	Timestamp time.Time `json:"timestamp"`
}

type GenerateProgramRequest struct {
	TrainingClass     string   `json:"training_class"`
	Contraindications []string `json:"contraindications"`
	Goals             []string `json:"goals"`
	FitnessLevel      string   `json:"fitness_level"`
	AgeGroup          string   `json:"age_group"`
	Gender            string   `json:"gender"`
	HasInjury         bool     `json:"has_injury"`
	DurationWeeks     int      `json:"duration_weeks"`
}
