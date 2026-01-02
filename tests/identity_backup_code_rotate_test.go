package tests

import (
	"net/http"
	"testing"
)

func TestBackupCodeRotate(t *testing.T) {
	// Arrange
	adminAccessToken := adminToken(t)
	user := createUser(t, adminAccessToken)
	loginResp := login(t, user.Email, user.Password)
	payload := map[string]string{
		"current_password": user.Password,
	}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/mfa/backup_code/rotate", payload, loginResp.AccessToken)

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("backup code rotate failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		RecoveryCodes []string `json:"recovery_codes"`
	}
	decodeSuccess(t, body, &data)
	if len(data.RecoveryCodes) == 0 {
		t.Fatalf("expected recovery codes")
	}
}
