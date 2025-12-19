package hash

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHashVerify(t *testing.T) {
	cases := []struct {
		name     string
		password string
		pepper   string
		verify   string
		wantOK   bool
	}{
		{name: "match_plain", password: "secret123A", pepper: "", verify: "secret123A", wantOK: true},
		{name: "match_pepper", password: "secret123A", pepper: "pep", verify: "secret123A", wantOK: true},
		{name: "mismatch", password: "secret123A", pepper: "pep", verify: "other", wantOK: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hasher := NewBcrypt(bcrypt.MinCost, tc.pepper)
			hash, err := hasher.Hash(tc.password)
			if err != nil {
				t.Fatalf("Hash() error: %v", err)
			}
			if got := hasher.Verify(string(hash), tc.verify); got != tc.wantOK {
				t.Fatalf("Verify() = %v, want %v", got, tc.wantOK)
			}
		})
	}
}
