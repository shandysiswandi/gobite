package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
)

func TestLogin2FAWithBackupCode(t *testing.T) {
	t.Parallel()
	email := "mfa-backup@example.com"
	password := "ValidPassword123!"
	userID := int64(1101)
	factorID := int64(2101)
	codeID := int64(3101)
	backupCode := "backup-12345"

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := seedMFAFactorWithType(context.Background(), env.pool, factorID, userID, entity.MFATypeBackupCode, "Backup Codes", []byte(""), 1, true); err != nil {
		t.Fatalf("seed mfa factor: %v", err)
	}
	if err := seedMFABackupCode(context.Background(), env.pool, codeID, userID, backupCode); err != nil {
		t.Fatalf("seed backup code: %v", err)
	}

	status, body := doLoginRaw(t, email, password)
	if status != http.StatusOK {
		t.Fatalf("login failed: %s", body)
	}

	var loginResp struct {
		Data struct {
			MfaRequired    bool   `json:"mfa_required"`
			ChallengeToken string `json:"challenge_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &loginResp); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if !loginResp.Data.MfaRequired || loginResp.Data.ChallengeToken == "" {
		t.Fatalf("expected MFA challenge")
	}

	status, body = doJSONRequest(t, http.MethodPost, "/api/v1/identity/login/2fa", map[string]string{
		"challenge_token": loginResp.Data.ChallengeToken,
		"method":          "BackupCode",
		"code":            backupCode,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("login 2fa failed: %s", body)
	}

	var parsed struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("unmarshal login 2fa response: %v", err)
	}
	if parsed.Data.AccessToken == "" || parsed.Data.RefreshToken == "" {
		t.Fatalf("expected access and refresh tokens")
	}
}
