package tests

import (
	"net/http"
	"strconv"
	"testing"
)

func TestUsersDetail(t *testing.T) {
	// Arrange
	token := adminToken(t)
	user := createUser(t, token)
	path := "/api/v1/identity/users/" + strconv.FormatInt(user.ID, 10)

	// Act
	status, body := doJSON(t, http.MethodGet, path, nil, token)

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("user detail failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		User struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	decodeSuccess(t, body, &data)
	if data.User.Email != user.Email {
		t.Fatalf("expected user email %q, got %q", user.Email, data.User.Email)
	}
}
