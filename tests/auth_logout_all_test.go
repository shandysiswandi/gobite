package tests

import (
	"context"
	"net/http"
	"testing"
)

func TestLogoutAllRevokesTokens(t *testing.T) {
	t.Parallel()
	email := "logout-all@example.com"
	password := "ValidPassword123!"
	userID := int64(1403)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	accessToken, refreshToken := loginTokens(t, email, password)

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/logout-all", nil, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", status, body)
	}

	status, body = doJSONRequest(t, http.MethodPost, "/api/v1/identity/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", status, body)
	}
}
