package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
)

func TestPasswordReset(t *testing.T) {
	t.Parallel()
	email := "reset@example.com"
	oldPassword := "ValidPassword123!"
	newPassword := "NewPassword123!"
	userID := int64(1302)
	challengeID := int64(2302)
	token := "reset-token-1302"

	if err := seedUser(context.Background(), env.pool, userID, email, oldPassword); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := seedChallenge(context.Background(), env.pool, challengeID, userID, entity.ChallengePurposePasswordForgotReset, token, time.Now().Add(time.Hour), nil); err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/password/reset", map[string]string{
		"challenge_token": token,
		"new_password":    newPassword,
	}, nil)
	if status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", status, body)
	}

	status, body = doLoginRaw(t, email, newPassword)
	if status != http.StatusOK {
		t.Fatalf("expected login with new password: %d %s", status, body)
	}
}
