package tests

import (
	"testing"
)

func TestLogin(t *testing.T) {

	t.Run("WithoutMFAEnable", func(t *testing.T) {

		// Arrange
		email := adminEmail
		password := adminPassword

		// Act
		resp := login(t, email, password)

		// Assert
		if resp.AccessToken == "" || resp.RefreshToken == "" {
			t.Fatalf("expected tokens in login response")
		}
	})

	t.Run("WithMFAEnable", func(t *testing.T) {

		// Arrange
		email := userEmail
		password := userPassword

		// Act
		resp := login(t, email, password)

		// Assert
		if resp.AccessToken != "" || resp.RefreshToken != "" {
			t.Fatalf("expected tokens not in login response")
		}
		if !resp.MfaRequired || resp.ChallengeToken == "" || len(resp.AvailableMethods) == 0 {
			t.Fatalf("expected mfa_required, challenge_token, and available_methods not empty")
		}
	})
}
