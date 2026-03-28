package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	assert.NotNil(t, s)
}

func TestHealthCheckRegistered(t *testing.T) {
	s := NewServer()
	defer s.Stop()

	// Проверяем, что health check зарегистрирован
	serviceInfo := s.GetServiceInfo()
	_, ok := serviceInfo["grpc.health.v1.Health"]
	assert.True(t, ok, "Health service should be registered")
}
