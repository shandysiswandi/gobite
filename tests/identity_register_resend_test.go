package tests

import (
	"net/http"
	"testing"
)

func TestRegisterResend(t *testing.T) {

	// Arrange
	email := uniqueEmail("real-resend")
	registerPayload := map[string]string{
		"email":     email,
		"password":  "Secret123!",
		"full_name": "Test User",
	}
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/register", registerPayload, "")
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("register failed: status=%d message=%q", status, errEnv.Message)
	}

	payload := map[string]string{"email": email}

	// Act
	status, body = doJSON(t, http.MethodPost, "/api/v1/identity/register/resend", payload, "")

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("register resend failed: status=%d message=%q", status, errEnv.Message)
	}
}
