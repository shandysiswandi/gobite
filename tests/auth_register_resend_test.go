package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
)

func TestRegisterResend(t *testing.T) {
	t.Parallel()
	email := "resend@example.com"
	password := "ValidPassword123!"
	userID := int64(1201)

	if err := seedUserWithStatus(context.Background(), env.pool, userID, email, password, entity.UserStatusUnverified, nil); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/register/resend", map[string]string{
		"email": email,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
}
