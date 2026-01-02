package tests

import (
	"context"
	"net/http"
	"testing"
)

func TestRefreshToken(t *testing.T) {
	t.Parallel()
	email := "refresh@example.com"
	password := "ValidPassword123!"
	userID := int64(1401)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, refreshToken := loginTokens(t, email, password)

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
}
