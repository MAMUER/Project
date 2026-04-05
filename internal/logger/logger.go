// internal/logger/logger.go
package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger структура для обёртки над zap.Logger
type Logger struct {
	*zap.Logger
	service string
}

// New создает новый логгер с именем сервиса
func New(service string) *Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.StacktraceKey = "stacktrace"
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	// Добавляем уровень логов из переменной окружения
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(lvl)); err == nil {
			cfg.Level = zap.NewAtomicLevelAt(level)
		}
	}

	logger, err := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		log.Fatal("failed to initialize logger", zap.Error(err))
	}

	return &Logger{
		Logger:  logger,
		service: service,
	}
}

// Service возвращает имя сервиса для логирования
func (l *Logger) Service() string {
	return l.service
}

// WithRequestID добавляет request_id к контексту логгера
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger:  l.With(zap.String("request_id", requestID)), // Используем встроенный метод
		service: l.service,                                   // Сохраняем сервис!
	}
}

// WithFields добавляет произвольные поля к контексту логгера
func (l *Logger) WithFields(fields ...zap.Field) *zap.Logger {
	return l.With(fields...)
}

// Sync гарантирует запись всех буферизированных логов
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
