package tests

import (
	"net/http"
	"strconv"
	"testing"
)

func TestUsersDelete(t *testing.T) {
	// Arrange
	token := adminToken(t)
	user := createUser(t, token)
	path := "/api/v1/identity/users/" + strconv.FormatInt(user.ID, 10)

	// Act
	status, body := doJSON(t, http.MethodDelete, path, nil, token)

	// Assert
	if status != http.StatusNoContent {
		errEnv := decodeError(t, body)
		t.Fatalf("user delete failed: status=%d message=%q", status, errEnv.Message)
	}
}
