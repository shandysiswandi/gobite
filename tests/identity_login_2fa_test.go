package tests

import (
	"testing"
)

func TestLogin2FA(t *testing.T) {

	t.Run("WithTOTP", func(t *testing.T) {

		// Arrange
		loginResp := login(t, userEmail, userPassword)

		if !loginResp.MfaRequired || loginResp.ChallengeToken == "" {
			t.Fatalf("expected MFA challenge on login")
		}

		payload := map[string]string{
			"challenge_token": loginResp.ChallengeToken,
			"method":          "TOTP",
			"code":            totpCode(t, userTOTPSecret),
		}

		// Act
		status, body := doJSON(t, "POST", "/api/v1/identity/login/2fa", payload, "")
		if status != 200 {
			errEnv := decodeError(t, body)
			t.Fatalf("login 2fa failed: status=%d message=%q", status, errEnv.Message)
		}

		// Assert
		var data loginData
		decodeSuccess(t, body, &data)
		if data.AccessToken == "" || data.RefreshToken == "" {
			t.Fatalf("expected tokens in login 2fa response")
		}
	})

	t.Run("WithBackupCode", func(t *testing.T) {

		// Arrange
		email := userEmail
		password := userPassword

		loginResp := login(t, email, password)

		if !loginResp.MfaRequired || loginResp.ChallengeToken == "" {
			t.Fatalf("expected MFA challenge on login")
		}

		payload := map[string]string{
			"challenge_token": loginResp.ChallengeToken,
			"method":          "BackupCode",
			"code":            userBackupCode, // userBCTwo
		}

		// Act
		status, body := doJSON(t, "POST", "/api/v1/identity/login/2fa", payload, "")
		if status != 200 {
			errEnv := decodeError(t, body)
			t.Fatalf("login 2fa failed: status=%d message=%q", status, errEnv.Message)
		}

		// Assert
		var data loginData
		decodeSuccess(t, body, &data)
		if data.AccessToken == "" || data.RefreshToken == "" {
			t.Fatalf("expected tokens in login 2fa response")
		}
	})
}
