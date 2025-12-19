package strcase

import "testing"

func TestToLowerSnake(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "lower", input: "already_snake", want: "already_snake"},
		{name: "camel", input: "HelloWorld", want: "hello_world"},
		{name: "acronym", input: "XMLHTTP", want: "x_m_l_h_t_t_p"},
		{name: "mixed", input: "userID", want: "user_i_d"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ToLowerSnake(tc.input); got != tc.want {
				t.Fatalf("ToLowerSnake(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
