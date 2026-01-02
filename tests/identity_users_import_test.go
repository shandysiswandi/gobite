package tests

import (
	"net/http"
	"testing"
)

func TestUsersImport(t *testing.T) {
	// Arrange
	token := adminToken(t)
	payload := []map[string]any{
		{
			"email":     uniqueEmail("real-import"),
			"password":  "Secret123!",
			"full_name": "Test User",
			"status":    2,
		},
	}

	// Act
	status, body := doJSON(t, http.MethodPost, "/api/v1/identity/users-import", payload, token)

	// Assert
	if status != http.StatusOK {
		errEnv := decodeError(t, body)
		t.Fatalf("users import failed: status=%d message=%q", status, errEnv.Message)
	}

	var data struct {
		Created int `json:"created"`
		Updated int `json:"updated"`
	}
	decodeSuccess(t, body, &data)
	if data.Created+data.Updated != 1 {
		t.Fatalf("expected one import result, got created=%d updated=%d", data.Created, data.Updated)
	}
}
