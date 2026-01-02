package tests

import (
	"net/http"
	"testing"
)

func TestLogout(t *testing.T) {

	// Arrange
	resp := login(t, adminEmail, adminPassword)
	payload := map[string]string{"refresh_token": resp.RefreshToken}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/logout", payload, "")

	// Assert
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("logout failed: status=%d message=%q", status, errEnv.Message)
	}
}
