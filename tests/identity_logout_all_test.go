package tests

import (
	"net/http"
	"testing"
)

func TestLogoutAll(t *testing.T) {

	// Arrange
	loginResp := login(t, adminEmail, adminPassword)

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/logout-all", nil, loginResp.AccessToken)

	// Assert
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("logout-all failed: status=%d message=%q", status, errEnv.Message)
	}
}
