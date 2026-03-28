package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoveServerHeader(t *testing.T) {
	handler := RemoveServerHeader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Header().Get("Server") != "" {
		t.Error("Server header should be removed")
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	headers := rec.Header()
	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Missing X-Content-Type-Options header")
	}
	if headers.Get("X-Frame-Options") != "DENY" {
		t.Error("Missing X-Frame-Options header")
	}
	if headers.Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("Missing X-XSS-Protection header")
	}
	if headers.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("Missing Referrer-Policy header")
	}
	if headers.Get("Content-Security-Policy") == "" {
		t.Error("Missing Content-Security-Policy header")
	}
}

func TestSecurityHeadersWithTLS(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.TLS = &tls.ConnectionState{} // simulate HTTPS
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Error("Missing HSTS header when TLS")
	}
}
