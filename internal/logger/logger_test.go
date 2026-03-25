package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNew(t *testing.T) {
	log := New("test-service")
	assert.NotNil(t, log)
	defer log.Sync()
}

func TestNewWithMultipleServices(t *testing.T) {
	serviceNames := []string{"auth", "biometric", "training", "gateway"}

	for _, name := range serviceNames {
		t.Run(name, func(t *testing.T) {
			log := New(name)
			assert.NotNil(t, log)
			defer log.Sync()

			core, recorded := observer.New(zap.InfoLevel)
			testLogger := &Logger{Logger: zap.New(core)}
			loggerWithService := testLogger.With(zap.String("service", name))
			loggerWithService.Info("service started")

			logs := recorded.All()
			require.Len(t, logs, 1)

			found := false
			for _, field := range logs[0].Context {
				if field.Key == "service" && field.String == name {
					found = true
					break
				}
			}
			assert.True(t, found, "service field not found")
		})
	}
}

func TestWithRequestID(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	log := &Logger{Logger: zap.New(core)}

	tests := []struct {
		name      string
		requestID string
	}{
		{"empty request ID", ""},
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000"},
		{"short ID", "abc123"},
		{"long ID", "very-long-request-id-with-many-characters-123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerWithID := log.WithRequestID(tt.requestID)
			loggerWithID.Info("test message")

			logs := recorded.All()
			require.GreaterOrEqual(t, len(logs), 1)

			lastLog := logs[len(logs)-1]
			assert.Equal(t, "test message", lastLog.Message)

			if tt.requestID != "" {
				found := false
				for _, field := range lastLog.Context {
					if field.Key == "request_id" && field.String == tt.requestID {
						found = true
						break
					}
				}
				assert.True(t, found, "request_id field not found")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	log := &Logger{Logger: zap.New(core)}

	levels := []struct {
		level   string
		logFunc func(msg string, fields ...zap.Field)
	}{
		{"debug", log.Debug},
		{"info", log.Info},
		{"warn", log.Warn},
		{"error", log.Error},
	}

	for _, lvl := range levels {
		t.Run(lvl.level, func(t *testing.T) {
			lvl.logFunc("test message")
			logs := recorded.All()
			require.NotEmpty(t, logs)

			lastLog := logs[len(logs)-1]
			assert.Equal(t, "test message", lastLog.Message)
		})
	}
}
