package tests

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	libotp "github.com/pquerna/otp"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
)

const (
	adminEmail    = "admin@gobite.com"
	adminPassword = "Secret123!"

	userEmail      = "user@gobite.com"
	userPassword   = "Secret123!"
	userTOTPSecret = "KFP6EBHKHTWE2PHK5GOK7K2ARBZWQDBV"
	userBackupCode = "odwh-23IX-j5Pl" // or "LZEG-lnN4-w8hQ"
)

type loginData struct {
	AccessToken      string   `json:"access_token"`
	RefreshToken     string   `json:"refresh_token"`
	MfaRequired      bool     `json:"mfa_required"`
	ChallengeToken   string   `json:"challenge_token"`
	AvailableMethods []string `json:"available_methods"`
}

func login(t *testing.T, email, password string) loginData {
	t.Helper()

	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/login", payload, "")
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("login failed: status=%d message=%q", status, errEnv.Message)
	}

	var data loginData
	decodeSuccess(t, body, &data)

	return data
}

func adminToken(t *testing.T) string {
	t.Helper()

	resp := login(t, adminEmail, adminPassword)
	if resp.AccessToken == "" {
		t.Fatal("missing admin access token")
	}

	return resp.AccessToken
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

type testUser struct {
	ID       int64
	Email    string
	Password string
	FullName string
}

func createUser(t *testing.T, token string) testUser {
	t.Helper()

	user := testUser{
		Email:    uniqueEmail("real-user"),
		Password: "Secret123!",
		FullName: "Test User",
	}

	payload := map[string]any{
		"email":     user.Email,
		"password":  user.Password,
		"full_name": user.FullName,
		"status":    entity.UserStatusActive,
	}

	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/users", payload, token)
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("create user failed: status=%d message=%q", status, errEnv.Message)
	}

	user.ID = lookupUserID(t, token, user.Email)

	return user
}

func lookupUserID(t *testing.T, token, email string) int64 {
	t.Helper()

	path := "/api/v1/identity/users?search=" + url.QueryEscape(email) + "&size=1&page=1"
	status, body := doJSON(t, http.MethodGet, path, nil, token)
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("lookup user failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		Users []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"users"`
	}
	decodeSuccess(t, body, &data)
	if len(data.Users) == 0 {
		t.Fatalf("lookup user returned no results for %q", email)
	}

	id, err := strconv.ParseInt(data.Users[0].ID, 10, 64)
	if err != nil {
		t.Fatalf("parse user id: %v", err)
	}

	return id
}

func totpCode(t *testing.T, key string) string {
	t.Helper()

	issuer := "GOBITE"
	period := uint(30)
	skew := uint(1)

	generator := otp.NewTOTP(issuer, period, skew, libotp.DigitsSix)
	code, err := generator.GenerateCode(key, time.Now())
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}

	return code
}

func setupTOTP(t *testing.T, token, password string) (string, string) {
	t.Helper()

	payload := map[string]string{
		"friendly_name":    "Test MFA",
		"current_password": password,
	}

	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/mfa/totp/setup", payload, token)
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("totp setup failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		ChallengeToken string `json:"challenge_token"`
		Key            string `json:"key"`
	}
	decodeSuccess(t, body, &data)
	if data.ChallengeToken == "" || data.Key == "" {
		t.Fatal("totp setup response missing fields")
	}

	return data.ChallengeToken, data.Key
}
