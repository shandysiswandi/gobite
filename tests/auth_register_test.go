package tests

import (
	"net/http"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	status, body := doJSONRequest(t, http.MethodPost, "/api/v1/identity/register", map[string]string{
		"email":     "new.user@example.com",
		"password":  "ValidPassword123!",
		"full_name": "New User",
	}, nil)
	if status != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", status, body)
	}
}
