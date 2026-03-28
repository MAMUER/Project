package main

import (
	"testing"
)

func TestTrainingServiceMain(t *testing.T) {
	t.Skip("Integration test - requires database and RabbitMQ")
}

func TestGeneratePlan(t *testing.T) {
	t.Skip("Integration test - requires database")
}

func TestCompleteWorkout(t *testing.T) {
	t.Skip("Integration test - requires database")
}
