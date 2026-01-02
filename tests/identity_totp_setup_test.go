package tests

import (
	"testing"
)

func TestTOTPSetup(t *testing.T) {

	// Arrange
	adminAccessToken := adminToken(t)
	user := createUser(t, adminAccessToken)
	loginResp := login(t, user.Email, user.Password)

	// Act
	challengeToken, _ := setupTOTP(t, loginResp.AccessToken, user.Password)

	// Assert
	if challengeToken == "" {
		t.Fatalf("expected challenge token and key")
	}
}
