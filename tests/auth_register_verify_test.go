package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
)

func TestRegisterVerify(t *testing.T) {
	t.Parallel()
	email := "verify@example.com"
	password := "ValidPassword123!"
	userID := int64(1202)
	challengeID := int64(2202)
	token := "verify-token-1202"

	if err := seedUserWithStatus(context.Background(), env.pool, userID, email, password, entity.UserStatusUnverified, nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := seedChallenge(context.Background(), env.pool, challengeID, userID, entity.ChallengePurposeRegisterVerify, token, time.Now().Add(time.Hour), nil); err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/register/verify", map[string]string{
		"challenge_token": token,
	}, nil)
	if status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", status, body)
	}

	statusValue, err := fetchUserStatus(context.Background(), env.pool, userID)
	if err != nil {
		t.Fatalf("fetch user status: %v", err)
	}
	if statusValue != entity.UserStatusActive {
		t.Fatalf("expected user status active, got %s", statusValue.String())
	}
}
