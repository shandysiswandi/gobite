package tests

import "testing"

func TestUsersCreate(t *testing.T) {
	// Arrange
	token := adminToken(t)

	// Act
	user := createUser(t, token)

	// Assert
	if user.ID == 0 {
		t.Fatalf("expected created user id")
	}
}
