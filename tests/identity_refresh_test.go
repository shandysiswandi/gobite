package tests

import (
	"net/http"
	"testing"
)

func TestRefreshToken(t *testing.T) {

	// Arrange
	resp := login(t, adminEmail, adminPassword)
	payload := map[string]string{"refresh_token": resp.RefreshToken}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/refresh", payload, "")

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("refresh failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	decodeSuccess(t, body, &data)
	if data.AccessToken == "" || data.RefreshToken == "" {
		t.Fatalf("expected tokens in refresh response")
	}
}
