package tests

import (
	"context"
	"net/http"
	"testing"
)

func TestPasswordForgot(t *testing.T) {
	t.Parallel()
	email := "forgot@example.com"
	password := "ValidPassword123!"
	userID := int64(1301)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/password/forgot", map[string]string{
		"email": email,
	}, nil)
	if status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", status, body)
	}
}
