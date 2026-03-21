package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

func Init(level string, service string) error {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stdout"}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	globalLogger = logger.With(zap.String("service", service))
	zap.ReplaceGlobals(globalLogger)
	return nil
}

func Get() *zap.Logger {
	if globalLogger == nil {
		return zap.L()
	}
	return globalLogger
}

func Sync() {
	if globalLogger != nil {
		globalLogger.Sync()
	}
}