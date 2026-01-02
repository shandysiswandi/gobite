package tests

import (
	"net/http"
	"testing"
)

func TestUsersList(t *testing.T) {
	// Arrange
	token := adminToken(t)

	// Act
	status, body := doJSON(t, http.MethodGet, "/api/v1/identity/users?size=10&page=1", nil, token)

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("users list failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		Users []struct {
			ID string `json:"id"`
		} `json:"users"`
	}
	decodeSuccess(t, body, &data)
	if len(data.Users) == 0 {
		t.Fatalf("expected users in list")
	}
}
