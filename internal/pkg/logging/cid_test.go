package logging

import (
	"context"
	"testing"
)

func TestCorrelationID(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{name: "missing", ctx: context.Background(), want: "[invalid_chain_id]"},
		{name: "set", ctx: SetCorrelationID(context.Background(), "cid-1"), want: "cid-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GetCorrelationID(tc.ctx); got != tc.want {
				t.Fatalf("GetCorrelationID() = %q, want %q", got, tc.want)
			}
		})
	}
}
