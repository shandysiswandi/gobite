package tests

import (
	"context"
	"net/http"
	"testing"
)

func TestPasswordChange(t *testing.T) {
	t.Parallel()
	email := "change@example.com"
	oldPassword := "ValidPassword123!"
	newPassword := "NewPassword321!"
	userID := int64(1303)

	if err := seedUser(context.Background(), env.pool, userID, email, oldPassword); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	accessToken, _ := loginTokens(t, email, oldPassword)

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/password/change", map[string]string{
		"current_password": oldPassword,
		"new_password":     newPassword,
	}, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", status, body)
	}

	status, body = doLoginRaw(t, email, newPassword)
	if status != http.StatusOK {
		t.Fatalf("expected login with new password: %d %s", status, body)
	}
}
