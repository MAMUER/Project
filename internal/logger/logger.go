package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger struct {
    *zap.Logger
    service string
}

func New(service string) *Logger {
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "ts"
    config.EncoderConfig.MessageKey = "message"
    config.EncoderConfig.LevelKey = "level"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    config.InitialFields = map[string]interface{}{
        "service": service,
    }

    logger, err := config.Build()
    if err != nil {
        logger = zap.NewNop()
    }
    return &Logger{
        Logger:  logger,
        service: service,
    }
}

func (l *Logger) WithRequestID(requestID string) *zap.Logger {
    return l.Logger.With(zap.String("request_id", requestID))
}

// Обертки для удобства с string полями
func (l *Logger) Info(msg string, fields ...zap.Field) {
    l.Logger.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
    l.Logger.Error(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
    l.Logger.Warn(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
    l.Logger.Debug(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
    l.Logger.Fatal(msg, fields...)
}