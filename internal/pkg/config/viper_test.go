package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestViperGetters(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := []byte(`int_value: 42
bool_value: true
float_value: 3.14
string_value: hello
binary_value: aGVsbG8=
invalid_binary: "!!!"
array_value: "a,b,c"
map_value: "k1:v1,k2:v2"
`)
	if err := os.WriteFile(cfgPath, cfg, 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	vc, err := NewViper(cfgPath)
	if err != nil {
		t.Fatalf("NewViper() error: %v", err)
	}

	cases := []struct {
		name string
		got  any
		want any
	}{
		{name: "int", got: vc.GetInt("int_value"), want: int64(42)},
		{name: "bool", got: vc.GetBool("bool_value"), want: true},
		{name: "float", got: vc.GetFloat("float_value"), want: 3.14},
		{name: "string", got: vc.GetString("string_value"), want: "hello"},
		{name: "binary", got: vc.GetBinary("binary_value"), want: []byte("hello")},
		{name: "binary_invalid", got: vc.GetBinary("invalid_binary"), want: []byte(nil)},
		{name: "array", got: vc.GetArray("array_value"), want: []string{"a", "b", "c"}},
		{name: "map", got: vc.GetMap("map_value"), want: map[string]string{"k1": "v1", "k2": "v2"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if !reflect.DeepEqual(tc.got, tc.want) {
				t.Fatalf("got %#v, want %#v", tc.got, tc.want)
			}
		})
	}
}
