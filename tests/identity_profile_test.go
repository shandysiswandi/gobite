package tests

import (
	"net/http"
	"testing"
)

func TestProfile(t *testing.T) {

	// Arrange
	loginResp := login(t, adminEmail, adminPassword)

	// Act
	status, body := doJSON(t, http.MethodGet, "/api/v1/identity/profile", nil, loginResp.AccessToken)

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("profile failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		Email string `json:"email"`
	}
	decodeSuccess(t, body, &data)
	if data.Email != adminEmail {
		t.Fatalf("expected profile email %q, got %q", adminEmail, data.Email)
	}
}
