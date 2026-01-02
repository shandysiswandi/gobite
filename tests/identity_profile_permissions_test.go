package tests

import (
	"net/http"
	"testing"
)

func TestProfilePermissions(t *testing.T) {

	// Arrange
	loginResp := login(t, adminEmail, adminPassword)

	// Act
	status, body := doJSON(t, http.MethodGet, "/api/v1/identity/profile/permissions", nil, loginResp.AccessToken)

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("profile permissions failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		Permissions map[string][]string `json:"permissions"`
	}
	decodeSuccess(t, body, &data)
	if data.Permissions == nil {
		t.Fatalf("expected permissions map")
	}
}
