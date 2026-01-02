package tests

import (
	"net/http"
	"testing"
)

func TestProfileUpdate(t *testing.T) {

	// Arrange
	loginResp := login(t, adminEmail, adminPassword)
	payload := map[string]string{"full_name": "Updated User"}

	// Act
	status, body := doJSON(t, http.MethodPut, "/api/v1/identity/profile", payload, loginResp.AccessToken)

	// Assert
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("profile update failed: status=%d message=%q", status, errEnv.Message)
	}

	status, body = doJSON(t, http.MethodGet, "/api/v1/identity/profile", nil, loginResp.AccessToken)
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("profile read failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		FullName string `json:"full_name"`
	}
	decodeSuccess(t, body, &data)
	if data.FullName != "Updated User" {
		t.Fatalf("expected updated full_name")
	}
}
