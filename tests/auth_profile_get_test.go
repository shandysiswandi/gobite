package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProfileGet(t *testing.T) {
	t.Parallel()
	email := "profile@example.com"
	password := "ValidPassword123!"
	userID := int64(1501)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	accessToken, _ := loginTokens(t, email, password)

	status, body := doJSONRequest(t, http.MethodGet, "/api/v1/identity/profile", nil, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}

	var profileResp struct {
		Data struct {
			Email    string `json:"email"`
			FullName string `json:"full_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &profileResp); err != nil {
		t.Fatalf("unmarshal profile response: %v", err)
	}
	if profileResp.Data.Email != email {
		t.Fatalf("expected email %s, got %s", email, profileResp.Data.Email)
	}
}
