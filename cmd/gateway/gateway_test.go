package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler(t *testing.T) {
	g := &gateway{}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	g.healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGatewayMain(t *testing.T) {
	t.Skip("Integration test - requires all services")
}

func TestRegisterHandler(t *testing.T) {
	t.Skip("Integration test - requires user service")
}

func TestLoginHandler(t *testing.T) {
	t.Skip("Integration test - requires user service")
}

func TestClassifyHandler(t *testing.T) {
	t.Skip("Integration test - requires ML classifier service")
}

func TestGenerateMLPlanHandler(t *testing.T) {
	t.Skip("Integration test - requires ML generator service")
}

// Test timeout handling
func TestMLServiceTimeout(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Test with non-existent service
	_, err := client.Post("http://localhost:9999/classify", "application/json", nil)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// Test helper functions
func TestPtrHelpers(t *testing.T) {
	v32 := ptrInt32(42)
	if *v32 != 42 {
		t.Errorf("Expected 42, got %d", *v32)
	}

	vs := ptrString("test")
	if *vs != "test" {
		t.Errorf("Expected 'test', got %s", *vs)
	}
}
