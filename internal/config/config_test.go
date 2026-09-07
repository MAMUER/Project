package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    CacheConfig
		wantError bool
	}{
		{
			name:      "valid config",
			config:    CacheConfig{Addr: "localhost:6379", Password: "secret", DB: 0},
			wantError: false,
		},
		{
			name:      "empty addr",
			config:    CacheConfig{Addr: "", Password: "secret", DB: 0},
			wantError: true,
		},
		{
			name:      "empty password is valid",
			config:    CacheConfig{Addr: "localhost:6379", Password: "", DB: 0},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestJWTConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    JWTConfig
		wantError bool
	}{
		{
			name:      "valid config",
			config:    JWTConfig{PrivateKeyPEM: "private", PublicKeyPEM: "public"},
			wantError: false,
		},
		{
			name:      "empty private key",
			config:    JWTConfig{PrivateKeyPEM: "", PublicKeyPEM: "public"},
			wantError: true,
		},
		{
			name:      "empty public key",
			config:    JWTConfig{PrivateKeyPEM: "private", PublicKeyPEM: ""},
			wantError: true,
		},
		{
			name:      "both empty",
			config:    JWTConfig{PrivateKeyPEM: "", PublicKeyPEM: ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    ServerConfig
		wantError bool
	}{
		{
			name:      "valid config",
			config:    ServerConfig{Addr: ":8080"},
			wantError: false,
		},
		{
			name:      "empty addr",
			config:    ServerConfig{Addr: ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadCacheConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("VALKEY_ADDR"))
		require.NoError(t, os.Unsetenv("VALKEY_PASSWORD"))
		require.NoError(t, os.Unsetenv("VALKEY_DB"))

		cfg := LoadCacheConfig()
		assert.Equal(t, "localhost:6379", cfg.Addr)
		assert.Empty(t, cfg.Password)
		assert.Equal(t, 0, cfg.DB)
	})

	t.Run("custom values", func(t *testing.T) {
		require.NoError(t, os.Setenv("VALKEY_ADDR", "valkey:6379"))
		require.NoError(t, os.Setenv("VALKEY_PASSWORD", "secret"))
		require.NoError(t, os.Setenv("VALKEY_DB", "1"))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("VALKEY_ADDR"))
			require.NoError(t, os.Unsetenv("VALKEY_PASSWORD"))
			require.NoError(t, os.Unsetenv("VALKEY_DB"))
		})

		cfg := LoadCacheConfig()
		assert.Equal(t, "valkey:6379", cfg.Addr)
		assert.Equal(t, "secret", cfg.Password)
		assert.Equal(t, 1, cfg.DB)
	})
}

func TestLoadJWTConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		require.NoError(t, os.Setenv("JWT_PRIVATE_KEY_PEM", "private-key"))
		require.NoError(t, os.Setenv("JWT_PUBLIC_KEY_PEM", "public-key"))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("JWT_PRIVATE_KEY_PEM"))
			require.NoError(t, os.Unsetenv("JWT_PUBLIC_KEY_PEM"))
		})

		cfg := LoadJWTConfig()
		assert.Equal(t, "private-key", cfg.PrivateKeyPEM)
		assert.Equal(t, "public-key", cfg.PublicKeyPEM)
	})

	t.Run("panics on missing private key", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("JWT_PRIVATE_KEY_PEM"))
		require.NoError(t, os.Unsetenv("JWT_PUBLIC_KEY_PEM"))
		assert.Panics(t, func() {
			LoadJWTConfig()
		})
	})
}

func TestLoadServerConfig(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("TEST_SERVER_ADDR"))
		cfg := LoadServerConfig("TEST_SERVER_ADDR", ":8080")
		assert.Equal(t, ":8080", cfg.Addr)
	})

	t.Run("custom value", func(t *testing.T) {
		require.NoError(t, os.Setenv("TEST_SERVER_ADDR", ":9090"))
		t.Cleanup(func() { require.NoError(t, os.Unsetenv("TEST_SERVER_ADDR")) })

		cfg := LoadServerConfig("TEST_SERVER_ADDR", ":8080")
		assert.Equal(t, ":9090", cfg.Addr)
	})
}

func TestRedactSecrets(t *testing.T) {
	t.Run("redacts cache password", func(t *testing.T) {
		cfg := CacheConfig{Addr: "localhost:6379", Password: "secret", DB: 0}
		redacted := redactSecrets(cfg)
		cacheCfg, ok := redacted.(struct {
			Addr     string
			Password string
			DB       int
		})
		require.True(t, ok)
		assert.Equal(t, "localhost:6379", cacheCfg.Addr)
		assert.Equal(t, "[REDACTED]", cacheCfg.Password)
		assert.Equal(t, 0, cacheCfg.DB)
	})

	t.Run("redacts JWT keys", func(t *testing.T) {
		cfg := JWTConfig{PrivateKeyPEM: "private", PublicKeyPEM: "public"}
		redacted := redactSecrets(cfg)
		jwtCfg, ok := redacted.(struct {
			PrivateKeyPEM string
			PublicKeyPEM  string
		})
		require.True(t, ok)
		assert.Equal(t, "[REDACTED]", jwtCfg.PrivateKeyPEM)
		assert.Equal(t, "[REDACTED]", jwtCfg.PublicKeyPEM)
	})

	t.Run("returns config as-is for unknown type", func(t *testing.T) {
		type CustomConfig struct {
			Value string
		}
		cfg := CustomConfig{Value: "test"}
		redacted := redactSecrets(cfg)
		assert.Equal(t, cfg, redacted)
	})
}

func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    DatabaseConfig
		wantError bool
	}{
		{
			name:      "valid config",
			config:    DatabaseConfig{Host: "localhost", Port: "5432", User: "postgres", Password: "secret", DBName: "mydb", SSLMode: "disable"},
			wantError: false,
		},
		{
			name:      "empty host",
			config:    DatabaseConfig{Port: "5432", User: "postgres", Password: "secret", DBName: "mydb"},
			wantError: true,
		},
		{
			name:      "empty password",
			config:    DatabaseConfig{Host: "localhost", Port: "5432", User: "postgres", DBName: "mydb"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{
			App:    "test",
			Server: ServerConfig{Addr: ":8080"},
			Database: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "secret",
				DBName:   "mydb",
				SSLMode:  "disable",
			},
			Cache: CacheConfig{Addr: "localhost:6379", DB: 0},
			JWT:   JWTConfig{PrivateKeyPEM: "private", PublicKeyPEM: "public"},
		}
		require.NoError(t, cfg.Validate())
	})

	t.Run("invalid server", func(t *testing.T) {
		cfg := Config{Server: ServerConfig{Addr: ""}}
		require.Error(t, cfg.Validate())
	})

	t.Run("invalid database", func(t *testing.T) {
		cfg := Config{Database: DatabaseConfig{}}
		require.Error(t, cfg.Validate())
	})
}

func TestLoadConfigFromViper(t *testing.T) {
	v := viper.New()
	v.Set("app.name", "testapp")
	v.Set("server.addr", ":9090")
	v.Set("database.host", "db.example.com")
	v.Set("database.port", "5433")
	v.Set("database.user", "admin")
	v.Set("database.password", "s3cret")
	v.Set("database.dbname", "appdb")
	v.Set("database.sslmode", "require")
	v.Set("cache.addr", "redis:6379")
	v.Set("cache.password", "redispass")
	v.Set("cache.db", 2)
	v.Set("jwt.private_key_pem", "priv")
	v.Set("jwt.public_key_pem", "pub")

	cfg := LoadConfig(v)
	assert.Equal(t, "testapp", cfg.App)
	assert.Equal(t, ":9090", cfg.Server.Addr)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "redis:6379", cfg.Cache.Addr)
	assert.Equal(t, 2, cfg.Cache.DB)
}

func TestViperHelpers(t *testing.T) {
	v := viper.New()
	v.Set("existing_str", "hello")
	v.Set("existing_int", 42)
	v.Set("existing_bool", true)
	v.Set("existing_duration", "5s")
	v.Set("existing_float", 3.14)

	assert.Equal(t, "hello", GetViperString(v, "existing_str", "default"))
	assert.Equal(t, "default", GetViperString(v, "missing_str", "default"))
	assert.Equal(t, 42, GetViperInt(v, "existing_int", 0))
	assert.Equal(t, 0, GetViperInt(v, "missing_int", 0))
	assert.Equal(t, true, GetViperBool(v, "existing_bool", false))
	assert.Equal(t, false, GetViperBool(v, "missing_bool", false))
	assert.Equal(t, 5*time.Second, GetViperDuration(v, "existing_duration", "1s"))
	assert.Equal(t, 1*time.Second, GetViperDuration(v, "missing_duration", "1s"))
	assert.Equal(t, 3.14, GetViperFloat64(v, "existing_float", 0.0))
	assert.Equal(t, 0.0, GetViperFloat64(v, "missing_float", 0.0))
}
