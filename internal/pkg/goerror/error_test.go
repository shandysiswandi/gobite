package goerror

import (
	"errors"
	"net/http"
	"testing"
)

func TestType_String(t *testing.T) {
	cases := []struct {
		name string
		in   Type
		want string
	}{
		{name: "server", in: TypeServer, want: "ERROR_TYPE_SERVER"},
		{name: "business", in: TypeBusiness, want: "ERROR_TYPE_BUSINESS"},
		{name: "validation", in: TypeValidation, want: "ERROR_TYPE_VALIDATION"},
		{name: "unknown", in: Type(99), want: "ERROR_TYPE_UNKNOWN"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.in.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCode_String(t *testing.T) {
	cases := []struct {
		name string
		in   Code
		want string
	}{
		{name: "internal", in: CodeInternal, want: "ERROR_CODE_INTERNAL"},
		{name: "invalid_format", in: CodeInvalidFormat, want: "ERROR_CODE_INVALID_FORMAT"},
		{name: "invalid_input", in: CodeInvalidInput, want: "ERROR_CODE_INVALID_INPUT"},
		{name: "not_found", in: CodeNotFound, want: "ERROR_CODE_NOT_FOUND"},
		{name: "conflict", in: CodeConflict, want: "ERROR_CODE_CONFLICT"},
		{name: "unauthorized", in: CodeUnauthorized, want: "ERROR_CODE_UNAUTHORIZED"},
		{name: "forbidden", in: CodeForbidden, want: "ERROR_CODE_FORBIDDEN"},
		{name: "unknown", in: Code(99), want: "ERROR_CODE_INTERNAL"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.in.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNewErrorConstructors(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		check  func(t *testing.T, err error)
		msg    string
		errTyp Type
		code   Code
	}{
		{
			name:   "new_server",
			err:    NewServer(errors.New("db down")),
			msg:    "Internal server error",
			errTyp: TypeServer,
			code:   CodeInternal,
		},
		{
			name:   "new_business",
			err:    NewBusiness("rule", CodeConflict),
			msg:    "rule",
			errTyp: TypeBusiness,
			code:   CodeConflict,
		},
		{
			name:   "new_invalid_input",
			err:    NewInvalidInput(errors.New("bad input")),
			msg:    "validation error",
			errTyp: TypeValidation,
			code:   CodeInvalidInput,
		},
		{
			name:   "new_invalid_format",
			err:    NewInvalidFormat(),
			msg:    "invalid request body",
			errTyp: TypeValidation,
			code:   CodeInvalidFormat,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var perr *Error
			if !errors.As(tc.err, &perr) {
				t.Fatalf("expected *Error, got %T", tc.err)
			}
			if perr.Msg() != tc.msg {
				t.Fatalf("Msg() = %q, want %q", perr.Msg(), tc.msg)
			}
			if perr.Type() != tc.errTyp {
				t.Fatalf("Type() = %v, want %v", perr.Type(), tc.errTyp)
			}
			if perr.Code() != tc.code {
				t.Fatalf("Code() = %v, want %v", perr.Code(), tc.code)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	cases := []struct {
		name string
		in   *Error
		want string
	}{
		{
			name: "underlying_error",
			in:   &Error{err: errors.New("boom"), msg: "ignored", errType: TypeBusiness},
			want: "boom",
		},
		{
			name: "message",
			in:   &Error{msg: "custom", errType: TypeBusiness},
			want: "custom",
		},
		{
			name: "validation_default",
			in:   &Error{errType: TypeValidation},
			want: "Validation violation",
		},
		{
			name: "business_default",
			in:   &Error{errType: TypeBusiness},
			want: "Logical business not meet with requirement",
		},
		{
			name: "server_default",
			in:   &Error{errType: TypeServer},
			want: "Internal error",
		},
		{
			name: "unknown_default",
			in:   &Error{errType: Type(99)},
			want: "Unknown error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.in.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestError_String(t *testing.T) {
	err := &Error{
		err:     errors.New("boom"),
		msg:     "oops",
		errType: TypeBusiness,
		code:    CodeConflict,
	}
	want := "Error Type: ERROR_TYPE_BUSINESS, Code: ERROR_CODE_CONFLICT, Message: oops, Underlying Error: boom"
	if got := err.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestError_Unwrap(t *testing.T) {
	base := errors.New("root")
	err := &Error{err: base}
	if got := err.Unwrap(); got != base {
		t.Fatalf("Unwrap() = %v, want %v", got, base)
	}
	if !errors.Is(err, base) {
		t.Fatalf("errors.Is() = false, want true")
	}
}

func TestError_StatusCode(t *testing.T) {
	cases := []struct {
		name string
		in   Code
		want int
	}{
		{name: "invalid_format", in: CodeInvalidFormat, want: http.StatusBadRequest},
		{name: "invalid_input", in: CodeInvalidInput, want: http.StatusUnprocessableEntity},
		{name: "not_found", in: CodeNotFound, want: http.StatusNotFound},
		{name: "unauthorized", in: CodeUnauthorized, want: http.StatusUnauthorized},
		{name: "forbidden", in: CodeForbidden, want: http.StatusForbidden},
		{name: "timeout", in: CodeTimeout, want: http.StatusRequestTimeout},
		{name: "conflict", in: CodeConflict, want: http.StatusConflict},
		{name: "internal", in: CodeInternal, want: http.StatusInternalServerError},
		{name: "unknown", in: Code(99), want: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := &Error{errType: TypeServer, code: tc.in}
			if got := err.StatusCode(); got != tc.want {
				t.Fatalf("StatusCode() = %d, want %d", got, tc.want)
			}
		})
	}
}
