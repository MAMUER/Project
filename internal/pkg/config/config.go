package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	// Service
	ServiceName string `mapstructure:"SERVICE_NAME"`
	ServicePort int    `mapstructure:"SERVICE_PORT"`

	// Database
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     int    `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSSLMode  string `mapstructure:"DB_SSL_MODE"`

	// Redis
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// RabbitMQ
	RabbitMQURL string `mapstructure:"RABBITMQ_URL"`

	// JWT
	JWTSecret string `mapstructure:"JWT_SECRET"`
	JWTExpiry int    `mapstructure:"JWT_EXPIRY"`

	// gRPC
	MLClassificationAddr string `mapstructure:"ML_CLASSIFICATION_ADDR"`
	MLGenerationAddr     string `mapstructure:"ML_GENERATION_ADDR"`
	JavaLegacyAddr       string `mapstructure:"JAVA_LEGACY_ADDR"`

	// Logging
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

func Load(serviceName string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/healthfit/")

	// Переменные окружения имеют приоритет
	viper.AutomaticEnv()

	// Значения по умолчанию
	viper.SetDefault("SERVICE_NAME", serviceName)
	viper.SetDefault("SERVICE_PORT", 8080)
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("JWT_EXPIRY", 86400) // 24 часа
	viper.SetDefault("LOG_LEVEL", "info")

	if err := viper.ReadInConfig(); err != nil {
		// Конфиг не обязателен, используем переменные окружения
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
