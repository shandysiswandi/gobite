package tests

import (
	"net/http"
	"testing"
)

func TestUsersExport(t *testing.T) {
	// Arrange
	token := adminToken(t)

	// Act
	status, body := doJSON(t, http.MethodGet, "/api/v1/identity/users-export", nil, token)

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("users export failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		Users []struct {
			ID string `json:"id"`
		} `json:"users"`
	}
	decodeSuccess(t, body, &data)
	if data.Users == nil {
		t.Fatalf("expected users export data")
	}
}
