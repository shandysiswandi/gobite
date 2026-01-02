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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
)

const (
	testHMACSecret   = "test-hmac-secret"
	testArgon2Pepper = "pepper"
)

func doJSONRequest(t *testing.T, method, path string, body any, headers map[string]string) (int, string) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, env.baseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return resp.StatusCode, strings.TrimSpace(string(respBody))
}

func loginTokens(t *testing.T, email, password string) (string, string) {
	t.Helper()

	status, body := doLoginRaw(t, email, password)
	if status != http.StatusOK {
		t.Fatalf("login failed: %s", body)
	}

	var parsed struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}

	if parsed.Data.AccessToken == "" || parsed.Data.RefreshToken == "" {
		t.Fatalf("expected access and refresh tokens")
	}

	return parsed.Data.AccessToken, parsed.Data.RefreshToken
}

func doLogin(t *testing.T, email, password string) (int, string) {
	t.Helper()

	status, body := doLoginRaw(t, email, password)
	if status == http.StatusOK {
		return status, ""
	}

	var parsed struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err == nil && parsed.Message != "" {
		return status, parsed.Message
	}

	return status, body
}

func doLoginRaw(t *testing.T, email, password string) (int, string) {
	t.Helper()

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

	return resp.StatusCode, strings.TrimSpace(string(body))
}

func seedChallenge(ctx context.Context, pool *pgxpool.Pool, challengeID, userID int64, purpose entity.ChallengePurpose, token string, expiresAt time.Time, metadata map[string]any) error {
	hmac := hash.NewHMACSHA256(testHMACSecret)
	tokenHash, err := hmac.Hash(token)
	if err != nil {
		return err
	}

	meta := []byte("{}")
	if metadata != nil {
		meta, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO auth_challenges (id, token, user_id, purpose, expires_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, challengeID, string(tokenHash), userID, purpose, expiresAt, meta)
	return err
}

func seedMFAFactorWithType(ctx context.Context, pool *pgxpool.Pool, factorID, userID int64, mfaType entity.MFAType, friendlyName string, secret []byte, keyVersion int16, verified bool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO auth_mfa_factors (id, user_id, type, friendly_name, secret, key_version, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, factorID, userID, mfaType, friendlyName, secret, keyVersion, verified)
	return err
}

func seedMFABackupCode(ctx context.Context, pool *pgxpool.Pool, codeID, userID int64, code string) error {
	argon := hash.NewArgon2id(testArgon2Pepper)
	hashed, err := argon.Hash(code)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO auth_mfa_backup_codes (id, user_id, code)
		VALUES ($1, $2, $3)
	`, codeID, userID, string(hashed))
	return err
}

func fetchUserStatus(ctx context.Context, pool *pgxpool.Pool, userID int64) (entity.UserStatus, error) {
	var status entity.UserStatus
	err := pool.QueryRow(ctx, `SELECT status FROM auth_users WHERE id = $1`, userID).Scan(&status)
	return status, err
}

func fetchUserFullName(ctx context.Context, pool *pgxpool.Pool, userID int64) (string, error) {
	var fullName string
	err := pool.QueryRow(ctx, `SELECT full_name FROM auth_users WHERE id = $1`, userID).Scan(&fullName)
	return fullName, err
}

func seedUser(ctx context.Context, pool *pgxpool.Pool, userID int64, email, password string) error {
	return seedUserWithStatus(ctx, pool, userID, email, password, entity.UserStatusActive, nil)
}

func seedUserWithStatus(ctx context.Context, pool *pgxpool.Pool, userID int64, email, password string, status entity.UserStatus, deletedAt *time.Time) error {
	bcrypt := hash.NewBcrypt(4, "pepper")
	hashed, err := bcrypt.Hash(password)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO auth_users (id, email, full_name, avatar_url, status, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, email, "Jane Doe", "https://example.com/avatar.png", status, deletedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO auth_user_credentials (user_id, password)
		VALUES ($1, $2)
	`, userID, string(hashed))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func seedMFAFactor(ctx context.Context, pool *pgxpool.Pool, factorID, userID int64, verified bool) error {
	return seedMFAFactorWithType(ctx, pool, factorID, userID, entity.MFATypeTOTP, "Auth App", []byte("secret"), 1, verified)
}
