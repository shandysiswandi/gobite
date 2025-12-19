package uid

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUIDGenerate(t *testing.T) {
	cases := []struct {
		name string
		gen  *UUID
	}{
		{name: "uuid_v7", gen: NewUUID()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			id := tc.gen.Generate()
			parsed, err := uuid.Parse(id)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			if parsed.Version() != 7 {
				t.Fatalf("version = %d, want 7", parsed.Version())
			}
		})
	}
}

func TestSnowflakeGenerate(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "unique_ids"},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gen, err := NewSnowflake()
			if err != nil {
				t.Fatalf("NewSnowflake() error: %v", err)
			}

			id1 := gen.Generate()
			id2 := gen.Generate()
			if id1 == id2 {
				t.Fatalf("expected unique ids, got %d", id1)
			}
		})
	}
}
