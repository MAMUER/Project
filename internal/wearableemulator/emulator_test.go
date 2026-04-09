package wearableemulator

import (
	"math"
	"testing"
	"time"
)

func TestGetDefaultCapabilities(t *testing.T) {
	tests := []struct {
		deviceType    DeviceType
		expectedHR    bool
		expectedECG   bool
		expectedBP    bool
		expectedSpO2  bool
		expectedTemp  bool
		expectedSleep bool
		expectedSteps bool
		expectedHRV   bool
	}{
		{AppleWatch, true, true, false, true, false, true, true, true},
		{SamsungGalaxyWatch, true, true, false, true, true, true, true, true},
		{HuaweiWatchD2, true, true, true, true, true, true, true, true},
		{AmazfitTRex3, true, false, false, true, false, true, true, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.deviceType), func(t *testing.T) {
			caps := GetDefaultCapabilities(tt.deviceType)
			if caps.HeartRate != tt.expectedHR {
				t.Errorf("HeartRate = %v, want %v", caps.HeartRate, tt.expectedHR)
			}
			if caps.ECG != tt.expectedECG {
				t.Errorf("ECG = %v, want %v", caps.ECG, tt.expectedECG)
			}
			if caps.BloodPressure != tt.expectedBP {
				t.Errorf("BloodPressure = %v, want %v", caps.BloodPressure, tt.expectedBP)
			}
			if caps.SpO2 != tt.expectedSpO2 {
				t.Errorf("SpO2 = %v, want %v", caps.SpO2, tt.expectedSpO2)
			}
			if caps.Temperature != tt.expectedTemp {
				t.Errorf("Temperature = %v, want %v", caps.Temperature, tt.expectedTemp)
			}
			if caps.Sleep != tt.expectedSleep {
				t.Errorf("Sleep = %v, want %v", caps.Sleep, tt.expectedSleep)
			}
			if caps.Steps != tt.expectedSteps {
				t.Errorf("Steps = %v, want %v", caps.Steps, tt.expectedSteps)
			}
			if caps.HRV != tt.expectedHRV {
				t.Errorf("HRV = %v, want %v", caps.HRV, tt.expectedHRV)
			}
		})
	}
}

func TestDataGeneratorGenerateHeartRate(t *testing.T) {
	state := DefaultPhysiologicalState()
	gen := NewDataGenerator(state)

	// Generate multiple heart rates and check they're in valid range
	for i := 0; i < 100; i++ {
		hr := gen.GenerateHeartRate()
		if hr < 40 || hr > 200 {
			t.Errorf("Heart rate %f out of range [40, 200]", hr)
		}
	}
}

func TestDataGeneratorGenerateSpO2(t *testing.T) {
	state := DefaultPhysiologicalState()
	gen := NewDataGenerator(state)

	for i := 0; i < 100; i++ {
		spo2 := gen.GenerateSpO2()
		if spo2 < 90 || spo2 > 100 {
			t.Errorf("SpO2 %f out of range [90, 100]", spo2)
		}
	}
}

func TestDataGeneratorGenerateTemperature(t *testing.T) {
	state := DefaultPhysiologicalState()
	gen := NewDataGenerator(state)

	for i := 0; i < 100; i++ {
		temp := gen.GenerateTemperature()
		if temp < 35.5 || temp > 38.5 {
			t.Errorf("Temperature %f out of range [35.5, 38.5]", temp)
		}
	}
}

func TestDataGeneratorGenerateBloodPressure(t *testing.T) {
	state := DefaultPhysiologicalState()
	gen := NewDataGenerator(state)

	for i := 0; i < 100; i++ {
		sys, dia := gen.GenerateBloodPressure()
		if sys < 80 || sys > 200 {
			t.Errorf("Systolic %f out of range [80, 200]", sys)
		}
		if dia < 50 || dia > 130 {
			t.Errorf("Diastolic %f out of range [50, 130]", dia)
		}
		if sys <= dia {
			t.Errorf("Systolic %f should be > Diastolic %f", sys, dia)
		}
	}
}

func TestDataGeneratorGenerateHRV(t *testing.T) {
	state := DefaultPhysiologicalState()
	gen := NewDataGenerator(state)

	for i := 0; i < 100; i++ {
		hrv := gen.GenerateHRV()
		if hrv < 10 || hrv > 100 {
			t.Errorf("HRV %f out of range [10, 100]", hrv)
		}
	}
}

func TestDataGeneratorGenerateSleepStage(t *testing.T) {
	state := DefaultPhysiologicalState()
	gen := NewDataGenerator(state)

	validStages := map[string]bool{
		"deep":  true,
		"light": true,
		"rem":   true,
		"awake": true,
	}

	for i := 0; i < 50; i++ {
		stage := gen.GenerateSleepStage()
		if !validStages[stage] {
			t.Errorf("Invalid sleep stage: %s", stage)
		}
	}
}

func TestDataGeneratorGenerateSteps(t *testing.T) {
	state := DefaultPhysiologicalState()
	gen := NewDataGenerator(state)

	for i := 0; i < 100; i++ {
		steps := gen.GenerateSteps()
		if steps < 0 {
			t.Errorf("Steps %f should be >= 0", steps)
		}
	}
}

func TestNewDeviceEmulator(t *testing.T) {
	state := DefaultPhysiologicalState()
	emulator := NewDeviceEmulator("test-device-1", "user-1", "token-1", AppleWatch, state)

	if emulator.DeviceID != "test-device-1" {
		t.Errorf("DeviceID = %s, want test-device-1", emulator.DeviceID)
	}
	if emulator.UserID != "user-1" {
		t.Errorf("UserID = %s, want user-1", emulator.UserID)
	}
	if emulator.DeviceType != AppleWatch {
		t.Errorf("DeviceType = %s, want apple_watch", emulator.DeviceType)
	}
	if !emulator.Capabilities.HeartRate {
		t.Error("Apple Watch should have heart rate capability")
	}
	if !emulator.Capabilities.ECG {
		t.Error("Apple Watch should have ECG capability")
	}
}

func TestDeviceEmulatorGenerateBatch(t *testing.T) {
	state := DefaultPhysiologicalState()
	emulator := NewDeviceEmulator("test-device-2", "user-2", "token-2", HuaweiWatchD2, state)

	batch := emulator.GenerateBatch()
	if len(batch) == 0 {
		t.Error("Generated batch should not be empty")
	}

	// Check all samples have valid fields
	for _, sample := range batch {
		if sample.DeviceType == "" {
			t.Error("Sample device_type should not be empty")
		}
		if sample.MetricType == "" {
			t.Error("Sample metric_type should not be empty")
		}
		if sample.Quality == "" {
			t.Error("Sample quality should not be empty")
		}
	}
}

func TestSyncManagerRegisterUnregister(t *testing.T) {
	sm := NewSyncManager()
	state := DefaultPhysiologicalState()
	emulator := NewDeviceEmulator("test-device-3", "user-3", "token-3", AppleWatch, state)

	sm.RegisterDevice(emulator)

	retrieved, exists := sm.GetEmulator("test-device-3")
	if !exists {
		t.Error("Device should exist after registration")
	}
	if retrieved.DeviceID != emulator.DeviceID {
		t.Errorf("Retrieved device ID = %s, want %s", retrieved.DeviceID, emulator.DeviceID)
	}

	sm.UnregisterDevice("test-device-3")
	_, exists = sm.GetEmulator("test-device-3")
	if exists {
		t.Error("Device should not exist after unregistration")
	}
}

func TestSyncAllDevices(t *testing.T) {
	sm := NewSyncManager()
	state := DefaultPhysiologicalState()

	emulator1 := NewDeviceEmulator("test-device-4", "user-4", "token-4", AppleWatch, state)
	emulator1.SyncInterval = 1 * time.Millisecond // Force sync

	sm.RegisterDevice(emulator1)

	// Wait for sync interval to pass
	time.Sleep(10 * time.Millisecond)

	results := sm.SyncAllDevices()
	if len(results) == 0 {
		t.Error("Should have synced at least one device")
	}

	if _, ok := results["test-device-4"]; !ok {
		t.Error("Should have synced test-device-4")
	}
}

func TestUserPhysiologicalStateCalculateBMI(t *testing.T) {
	tests := []struct {
		weight float64
		height int
		want   float64
	}{
		{75, 175, 24.49},
		{90, 180, 27.78},
		{60, 170, 20.76},
	}

	for _, tt := range tests {
		// BMI = weight / (height_m)^2
		heightM := float64(tt.height) / 100.0
		got := tt.weight / (heightM * heightM)
		if math.Abs(got-tt.want) > 0.1 {
			t.Errorf("BMI(%f, %d) = %f, want %f", tt.weight, tt.height, got, tt.want)
		}
	}
}

func TestUserPhysiologicalStateCalculateMaxHeartRate(t *testing.T) {
	tests := []struct {
		age  int
		want float64
	}{
		{30, 190},
		{40, 180},
		{50, 170},
		{20, 200},
	}

	for _, tt := range tests {
		// Max HR = 220 - age
		got := 220 - tt.age
		if float64(got) != tt.want {
			t.Errorf("MaxHR(age=%d) = %d, want %f", tt.age, got, tt.want)
		}
	}
}

func TestSleepStageToValue(t *testing.T) {
	tests := []struct {
		stage string
		want  float64
	}{
		{"deep", 4},
		{"light", 3},
		{"rem", 2},
		{"awake", 1},
		{"invalid", 0},
	}

	for _, tt := range tests {
		got := sleepStageToValue(tt.stage)
		if got != tt.want {
			t.Errorf("sleepStageToValue(%s) = %f, want %f", tt.stage, got, tt.want)
		}
	}
}

func TestHasHealthRisk(t *testing.T) {
	tests := []struct {
		name     string
		state    *UserPhysiologicalState
		wantRisk bool
	}{
		{
			name:     "Normal state",
			state:    &UserPhysiologicalState{BaseHeartRate: 72},
			wantRisk: false,
		},
		{
			name:     "High heart rate",
			state:    &UserPhysiologicalState{BaseHeartRate: 110},
			wantRisk: true,
		},
		{
			name:     "Low heart rate",
			state:    &UserPhysiologicalState{BaseHeartRate: 45},
			wantRisk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.BaseHeartRate > 100 || tt.state.BaseHeartRate < 50
			if got != tt.wantRisk {
				t.Errorf("hasHealthRisk() = %v, want %v", got, tt.wantRisk)
			}
		})
	}
}
