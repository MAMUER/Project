package auth

import (
	"testing"
)

const testSecret = "test-secret"

func TestSignResponse(t *testing.T) {
	secret := testSecret
	data := map[string]string{"key": "value"}
	sig, err := SignResponse(data, secret)
	if err != nil {
		t.Fatalf("SignResponse failed: %v", err)
	}
	if sig == "" {
		t.Error("Empty signature")
	}
}

func TestVerifyResponse(t *testing.T) {
	secret := testSecret
	data := map[string]string{"key": "value"}
	sig, _ := SignResponse(data, secret)

	if !VerifyResponse(data, sig, secret) {
		t.Error("VerifyResponse failed for valid signature")
	}
	if VerifyResponse(data, "invalid", secret) {
		t.Error("VerifyResponse passed for invalid signature")
	}
}
