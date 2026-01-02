package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTOTPSetup(t *testing.T) {
	t.Parallel()
	email := "totp@example.com"
	password := "ValidPassword123!"
	userID := int64(1601)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	accessToken, _ := loginTokens(t, email, password)

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/mfa/totp/setup", map[string]string{
		"friendly_name":    "Auth App",
		"current_password": password,
	}, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}

	var setupResp struct {
		Data struct {
			ChallengeToken string `json:"challenge_token"`
			Key            string `json:"key"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &setupResp); err != nil {
		t.Fatalf("unmarshal totp setup response: %v", err)
	}
	if setupResp.Data.ChallengeToken == "" || setupResp.Data.Key == "" {
		t.Fatalf("expected challenge token and key")
	}
}
