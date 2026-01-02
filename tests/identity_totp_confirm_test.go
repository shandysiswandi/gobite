package tests

import (
	"net/http"
	"testing"
)

func TestTOTPConfirm(t *testing.T) {

	// Arrange
	adminAccessToken := adminToken(t)
	user := createUser(t, adminAccessToken)
	loginResp := login(t, user.Email, user.Password)
	challengeToken, key := setupTOTP(t, loginResp.AccessToken, user.Password)

	payload := map[string]string{
		"challenge_token": challengeToken,
		"code":            totpCode(t, key),
	}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/mfa/totp/confirm", payload, loginResp.AccessToken)
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("totp confirm failed: status=%d message=%q", status, errEnv.Message)
	}

}
