package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestBackupCodeRotate(t *testing.T) {
	t.Parallel()
	email := "backup-rotate@example.com"
	password := "ValidPassword123!"
	userID := int64(1602)

	if err := seedUser(context.Background(), env.pool, userID, email, password); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	accessToken, _ := loginTokens(t, email, password)

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/mfa/backup_code/rotate", map[string]string{
		"current_password": password,
	}, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}

	var rotateResp struct {
		Data struct {
			RecoveryCodes []string `json:"recovery_codes"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &rotateResp); err != nil {
		t.Fatalf("unmarshal backup code response: %v", err)
	}
	if len(rotateResp.Data.RecoveryCodes) == 0 {
		t.Fatalf("expected recovery codes")
	}
}
