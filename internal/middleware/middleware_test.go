package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MAMUER/Project/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const testSecretMW = "test-secret-mw"

func TestRequestID(t *testing.T) {
	tests := []struct {
		name          string
		requestHeader string
		expectedID    string
	}{
		{
			name:          "no header - generates new ID",
			requestHeader: "",
			expectedID:    "",
		},
		{
			name:          "with header - uses provided ID",
			requestHeader: "test-id-123",
			expectedID:    "test-id-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestID := r.Context().Value(RequestIDKey)
				assert.NotNil(t, requestID)

				if tt.expectedID != "" {
					assert.Equal(t, tt.expectedID, requestID)
				} else {
					assert.NotEmpty(t, requestID)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
			if tt.requestHeader != "" {
				req.Header.Set("X-Request-ID", tt.requestHeader)
			}
			rr := httptest.NewRecorder()

			middleware := RequestID(handler)
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			responseID := rr.Header().Get("X-Request-ID")
			if tt.expectedID != "" {
				assert.Equal(t, tt.expectedID, responseID)
			} else {
				assert.NotEmpty(t, responseID)
			}
		})
	}
}

func TestRequestIDMultipleRequests(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(RequestIDKey).(string)
		w.Header().Set("X-Request-ID", requestID)
		w.WriteHeader(http.StatusOK)
	})
	middleware := RequestID(handler)

	ids := make([]string, 5)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
		ids[i] = rr.Header().Get("X-Request-ID")
	}

	seen := make(map[string]bool)
	for _, id := range ids {
		assert.False(t, seen[id], "Duplicate ID: %s", id)
		seen[id] = true
	}
}

func TestAuthMiddleware(t *testing.T) {
	secret := testSecretMW
	log := zap.NewNop()

	validToken, err := auth.GenerateJWT("user-123", "test@example.com", "client", secret, 24)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID string
		expectedRole   string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: "user-123",
			expectedRole:   "client",
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusNotFound, // изменено с 401 на 404
			expectedUserID: "",
			expectedRole:   "",
		},
		{
			name:           "invalid format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusNotFound, // изменено с 401 на 404
		},
		{
			name:           "wrong prefix",
			authHeader:     "Basic token",
			expectedStatus: http.StatusNotFound, // изменено с 401 на 404
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.string",
			expectedStatus: http.StatusNotFound, // изменено с 401 на 404
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + generateExpiredToken(secret),
			expectedStatus: http.StatusNotFound, // изменено с 401 на 404
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value(UserIDKey)
				role := r.Context().Value(RoleKey)

				if tt.expectedUserID != "" {
					assert.Equal(t, tt.expectedUserID, userID)
					assert.Equal(t, tt.expectedRole, role)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			middleware := AuthMiddleware(secret, log)(handler)
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func generateExpiredToken(secret string) string {
	claims := auth.Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "client",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddlewareWithContext(t *testing.T) {
	secret := testSecretMW
	log := zap.NewNop()

	validToken, err := auth.GenerateJWT("user-456", "test@example.com", "admin", secret, 24)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey).(string)
		role := r.Context().Value(RoleKey).(string)

		assert.Equal(t, "user-456", userID)
		assert.Equal(t, "admin", role)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()

	middleware := AuthMiddleware(secret, log)(handler)
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddlewareLogging(t *testing.T) {
	secret := testSecretMW
	core, recorded := observer.New(zap.DebugLevel)
	log := zap.New(core)

	invalidToken := "invalid.token.string"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)
	rr := httptest.NewRecorder()

	middleware := AuthMiddleware(secret, log)(handler)
	middleware.ServeHTTP(rr, req)

	logs := recorded.All()
	assert.Equal(t, http.StatusNotFound, rr.Code) // изменено с 401 на 404

	found := false
	for _, logEntry := range logs {
		if logEntry.Message == "Invalid token" {
			found = true
			break
		}
	}
	assert.True(t, found, "Token validation error not logged")
}
