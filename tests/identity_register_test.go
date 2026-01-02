package tests

import (
	"net/http"
	"testing"
)

func TestRegister(t *testing.T) {

	// Arrange
	payload := map[string]string{
		"email":     uniqueEmail("real-register"),
		"password":  "Secret123!",
		"full_name": "Test User",
	}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/register", payload, "")

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("register failed: status=%d message=%q", status, errEnv.Message)
	}
}
