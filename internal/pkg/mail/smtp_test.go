package mail

import (
	"context"
	"errors"
	"testing"
)

func TestNewSMTP(t *testing.T) {
	cases := []struct {
		name      string
		cfg       SMTPConfig
		wantErr   error
		checkAuth bool
		wantAuth  bool
	}{
		{name: "missing_host", cfg: SMTPConfig{Port: 25}, wantErr: ErrSMTPHostPortRequired},
		{name: "missing_port", cfg: SMTPConfig{Host: "localhost"}, wantErr: ErrSMTPHostPortRequired},
		{
			name:      "valid_no_auth",
			cfg:       SMTPConfig{Host: "localhost", Port: 25},
			wantErr:   nil,
			checkAuth: true,
			wantAuth:  false,
		},
		{
			name:      "with_auth",
			cfg:       SMTPConfig{Host: "localhost", Port: 25, Username: "user", Password: "pass"},
			wantErr:   nil,
			checkAuth: true,
			wantAuth:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sender, err := NewSMTP(tc.cfg)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tc.checkAuth {
					if (sender.auth != nil) != tc.wantAuth {
						t.Fatalf("auth configured = %t, want %t", sender.auth != nil, tc.wantAuth)
					}
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestSMTP_Send(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		cfg SMTPConfig
		// Named input parameters for target function.
		msg     Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSMTP(tt.cfg)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := s.Send(context.Background(), tt.msg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Send() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Send() succeeded unexpectedly")
			}
		})
	}
}
