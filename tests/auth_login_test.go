package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
)

func TestLoginWithoutMFA(t *testing.T) {
	t.Parallel()
	email := "jane.doe@example.com"
	password := "CorrectHorseBatteryStaple!"
	userID := int64(1001)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	reqBody, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, env.baseURL+"/api/v1/identity/login", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if parsed.Data.AccessToken == "" {
		t.Fatalf("expected access token")
	}
	if parsed.Data.RefreshToken == "" {
		t.Fatalf("expected refresh token")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	t.Parallel()
	email := "invalid.pass@example.com"
	password := "ValidPassword123!"
	userID := int64(1002)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	status, errMsg := doLogin(t, email, "WrongPassword!")
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", status, errMsg)
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	t.Parallel()
	status, errMsg := doLogin(t, "unknown@example.com", "whatever")
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", status, errMsg)
	}
}

func TestLoginUnverifiedUser(t *testing.T) {
	t.Parallel()
	email := "unverified@example.com"
	password := "ValidPassword123!"
	userID := int64(1003)

	if err := seedUserWithStatus(context.Background(), env.pool, userID, email, password, entity.UserStatusUnverified, nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	status, errMsg := doLogin(t, email, password)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", status, errMsg)
	}
}

func TestLoginBannedUser(t *testing.T) {
	t.Parallel()
	email := "banned@example.com"
	password := "ValidPassword123!"
	userID := int64(1004)

	if err := seedUserWithStatus(context.Background(), env.pool, userID, email, password, entity.UserStatusBanned, nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	status, errMsg := doLogin(t, email, password)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", status, errMsg)
	}
}

func TestLoginDeletedUser(t *testing.T) {
	t.Parallel()
	email := "deleted@example.com"
	password := "ValidPassword123!"
	userID := int64(1005)

	if err := seedUserWithStatus(context.Background(), env.pool, userID, email, password, entity.UserStatusDeleted, nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	status, errMsg := doLogin(t, email, password)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", status, errMsg)
	}
}

func TestLoginUnknownStatusUser(t *testing.T) {
	t.Parallel()
	email := "unknown-status@example.com"
	password := "ValidPassword123!"
	userID := int64(1006)

	if err := seedUserWithStatus(context.Background(), env.pool, userID, email, password, entity.UserStatusUnknown, nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	status, errMsg := doLogin(t, email, password)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", status, errMsg)
	}
}

func TestLoginWithMFARequired(t *testing.T) {
	t.Parallel()
	email := "mfa@example.com"
	password := "ValidPassword123!"
	userID := int64(1007)
	factorID := int64(2001)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := seedMFAFactor(context.Background(), env.pool, factorID, userID, true); err != nil {
		t.Fatalf("seed mfa: %v", err)
	}

	status, body := doLoginRaw(t, email, password)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}

	var parsed struct {
		Data struct {
			MfaRequired      bool     `json:"mfa_required"`
			ChallengeToken   string   `json:"challenge_token"`
			AvailableMethods []string `json:"available_methods"`
			AccessToken      string   `json:"access_token"`
			RefreshToken     string   `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !parsed.Data.MfaRequired {
		t.Fatalf("expected mfa_required=true")
	}
	if parsed.Data.ChallengeToken == "" {
		t.Fatalf("expected challenge token")
	}
	if len(parsed.Data.AvailableMethods) == 0 {
		t.Fatalf("expected available methods")
	}
	if parsed.Data.AccessToken != "" || parsed.Data.RefreshToken != "" {
		t.Fatalf("expected no tokens when MFA required")
	}
}
