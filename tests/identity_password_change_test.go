package tests

import (
	"net/http"
	"testing"
)

func TestPasswordChange(t *testing.T) {

	// Arrange
	loginResp := login(t, adminEmail, adminPassword)
	newPassword := "Secret123!1"
	payload := map[string]string{
		"current_password": adminPassword,
		"new_password":     newPassword,
	}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/password/change", payload, loginResp.AccessToken)

	// Assert
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("password change failed: status=%d message=%q", status, errEnv.Message)
	}

	loginNew := login(t, adminEmail, newPassword)
	if loginNew.AccessToken == "" {
		t.Fatalf("expected access token after password change")
	}

	revertPayload := map[string]string{
		"current_password": newPassword,
		"new_password":     adminPassword,
	}
	status, body = doJSON(t, http.MethodPost, "/api/v1/identity/password/change", revertPayload, loginNew.AccessToken)
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("password revert failed: status=%d message=%q", status, errEnv.Message)
	}
}
