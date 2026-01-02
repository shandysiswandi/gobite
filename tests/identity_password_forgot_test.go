package tests

import (
	"net/http"
	"testing"
)

func TestPasswordForgot(t *testing.T) {

	// Arrange
	payload := map[string]string{"email": adminEmail}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/password/forgot", payload, "")

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("password forgot failed: status=%d message=%q", status, errEnv.Message)
	}
}
