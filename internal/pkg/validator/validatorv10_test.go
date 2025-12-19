package validator

import (
	"strings"
	"testing"
)

type sample struct {
	FullName string `validate:"required,alphaspace"`
	Password string `validate:"required,password"`
}

func TestV10ValidatorValidate(t *testing.T) {
	validator, err := NewV10Validator()
	if err != nil {
		t.Fatalf("NewV10Validator() error: %v", err)
	}

	cases := []struct {
		name    string
		data    sample
		wantErr bool
		keys    []string
	}{
		{
			name:    "valid",
			data:    sample{FullName: "John Doe", Password: "StrongPass1"},
			wantErr: false,
		},
		{
			name:    "invalid_password",
			data:    sample{FullName: "John Doe", Password: "weak"},
			wantErr: true,
			keys:    []string{"password"},
		},
		{
			name:    "invalid_name",
			data:    sample{FullName: "John_Doe", Password: "StrongPass1"},
			wantErr: true,
			keys:    []string{"full_name"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Validate(tc.data)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				verr, ok := err.(V10ValidationError)
				if !ok {
					t.Fatalf("expected V10ValidationError, got %T", err)
				}
				for _, key := range tc.keys {
					msg, found := verr[key]
					if !found {
						t.Fatalf("missing key %q in error map", key)
					}
					if strings.TrimSpace(msg) == "" {
						t.Fatalf("empty message for key %q", key)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
