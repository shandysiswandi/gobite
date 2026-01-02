package tests

import (
	"net/http"
	"strconv"
	"testing"
)

func TestUsersUpdate(t *testing.T) {
	// Arrange
	token := adminToken(t)
	user := createUser(t, token)
	path := "/api/v1/identity/users/" + strconv.FormatInt(user.ID, 10)
	payload := map[string]any{
		"full_name": "Updated User",
	}

	// Act
	status, body := doJSON(t, http.MethodPut, path, payload, token)

	// Assert
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("user update failed: status=%d message=%q", status, errEnv.Message)
	}

	status, body = doJSON(t, http.MethodGet, path, nil, token)
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("user detail failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		User struct {
			FullName string `json:"full_name"`
		} `json:"user"`
	}
	decodeSuccess(t, body, &data)
	if data.User.FullName != "Updated User" {
		t.Fatalf("expected updated full_name")
	}
}
