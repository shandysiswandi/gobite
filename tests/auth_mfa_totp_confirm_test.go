package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	pquernaotp "github.com/pquerna/otp"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
)

func TestTOTPConfirm(t *testing.T) {
	t.Parallel()
	email := "totp-confirm@example.com"
	password := "ValidPassword123!"
	userID := int64(1603)

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

	totpClient := otp.NewTOTP("GOBITE", 30, 1, pquernaotp.DigitsSix)
	code, err := totpClient.GenerateCode(setupResp.Data.Key, time.Now())
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}

	status, body = doJSONRequest(t, http.MethodPost, "/api/v1/identity/mfa/totp/confirm", map[string]string{
		"challenge_token": setupResp.Data.ChallengeToken,
		"code":            code,
	}, map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}

	var confirmResp struct {
		Data struct {
			RecoveryCodes []string `json:"recovery_codes"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &confirmResp); err != nil {
		t.Fatalf("unmarshal totp confirm response: %v", err)
	}
	if len(confirmResp.Data.RecoveryCodes) == 0 {
		t.Fatalf("expected recovery codes")
	}
}
