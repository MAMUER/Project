package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger struct {
    *zap.Logger
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

    return &Logger{Logger: logger}
}

func (l *Logger) WithRequestID(requestID string) *zap.Logger {
    return l.Logger.With(zap.String("request_id", requestID))
}