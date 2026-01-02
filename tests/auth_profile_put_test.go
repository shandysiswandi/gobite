package tests

import (
	"context"
	"net/http"
	"testing"
)

func TestProfileUpdate(t *testing.T) {
	t.Parallel()
	email := "profile-update@example.com"
	password := "ValidPassword123!"
	userID := int64(1502)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	accessToken, _ := loginTokens(t, email, password)

	status, body := doJSONRequest(t, http.MethodPut, "/api/v1/identity/profile", map[string]string{
		"full_name": "Profile Updated",
	}, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", status, body)
	}

	fullName, err := fetchUserFullName(context.Background(), env.pool, userID)
	if err != nil {
		t.Fatalf("fetch user full name: %v", err)
	}
	if fullName != "Profile Updated" {
		t.Fatalf("expected updated full name, got %s", fullName)
	}
}
