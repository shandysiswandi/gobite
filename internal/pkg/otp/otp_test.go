package otp

import (
	"testing"
	"time"

	"github.com/pquerna/otp"
)

func TestTOTPGenerateValidate(t *testing.T) {
	cases := []struct {
		name       string
		digits     otp.Digits
		period     uint
		expectedLn int
	}{
		{name: "defaults_to_six", digits: otp.Digits(0), period: 0, expectedLn: 6},
		{name: "eight_digits", digits: otp.DigitsEight, period: 30, expectedLn: 8},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			totp := NewTOTP("gobite", tc.period, 1, tc.digits)

			secret, _, err := totp.Generate("user@example.com")
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			at := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
			code, err := totp.GenerateCode(secret, at)
			if err != nil {
				t.Fatalf("GenerateCode() error: %v", err)
			}
			if len(code) != tc.expectedLn {
				t.Fatalf("code length = %d, want %d", len(code), tc.expectedLn)
			}
			if !totp.Validate(code, secret, at) {
				t.Fatalf("Validate() = false, want true")
			}
		})
	}
}
