package main

import (
	"testing"
)

func TestBiometricServiceMain(t *testing.T) {
	t.Skip("Integration test - requires database and RabbitMQ")
}

func TestAddRecord(t *testing.T) {
	t.Skip("Integration test - requires database")
}

func TestBatchAddRecords(t *testing.T) {
	t.Skip("Integration test - requires database")
}
