package goerror

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrNotFound indicates that the requested resource could not be found.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict indicates that the request could not be completed due to a conflict.
	ErrConflict = errors.New("resource conflict")
)

// Type classifies errors into high-level buckets used by the application.
type Type int

const (
	// TypeServer represents server-side failures.
	TypeServer Type = iota
	// TypeBusiness represents business rule violations.
	TypeBusiness
	// TypeValidation represents input validation failures.
	TypeValidation
)

// String returns the string representation of the error type.
func (t Type) String() string {
	switch t {
	case TypeValidation:
		return "ERROR_TYPE_VALIDATION"
	case TypeBusiness:
		return "ERROR_TYPE_BUSINESS"
	case TypeServer:
		return "ERROR_TYPE_SERVER"
	default:
		return "ERROR_TYPE_UNKNOWN"
	}
}

// Code is a stable identifier used for mapping errors to HTTP status codes.
type Code int

const (
	// CodeInternal represents an internal or unspecified error.
	CodeInternal Code = iota
	// CodeInvalidFormat indicates invalid request format.
	CodeInvalidFormat
	// CodeInvalidInput indicates invalid request input.
	CodeInvalidInput
	// CodeNotFound indicates a missing resource.
	CodeNotFound
	// CodeConflict indicates a conflict (e.g., duplicate).
	CodeConflict
	// CodeTooManyRequest indicates rate limiting.
	CodeTooManyRequest
	// CodeUnauthorized indicates authentication failure.
	CodeUnauthorized
	// CodeForbidden indicates authorization failure.
	CodeForbidden
	// CodeTimeout indicates a timeout.
	CodeTimeout
)

// String returns the string representation of the error code.
func (c Code) String() string {
	switch c {
	case CodeInvalidFormat:
		return "ERROR_CODE_INVALID_FORMAT"
	case CodeInvalidInput:
		return "ERROR_CODE_INVALID_INPUT"
	case CodeNotFound:
		return "ERROR_CODE_NOT_FOUND"
	case CodeConflict:
		return "ERROR_CODE_CONFLICT"
	case CodeTooManyRequest:
		return "ERROR_CODE_TOO_MANY_REQUESTS"
	case CodeUnauthorized:
		return "ERROR_CODE_UNAUTHORIZED"
	case CodeForbidden:
		return "ERROR_CODE_FORBIDDEN"
	case CodeInternal:
		return "ERROR_CODE_INTERNAL"
	default:
		return "ERROR_CODE_INTERNAL"
	}
}

// Error is a structured error used across the application.
//
// It can wrap an underlying error while also carrying a user-facing message,
// a high-level type, and a stable error code.
type Error struct {
	err     error
	msg     string
	errType Type
	code    Code
	fields  map[string]string
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	if e.msg != "" {
		return e.msg
	}

	if e.errType == TypeValidation {
		return "Validation violation"
	}

	if e.errType == TypeBusiness {
		return "Logical business not meet with requirement"
	}

	if e.errType == TypeServer {
		return "Internal error"
	}

	return "Unknown error"
}

// String returns a verbose representation of the error for debugging/logging.
func (e *Error) String() string {
	return fmt.Sprintf(
		"Error Type: %s, Code: %s, Message: %s, Underlying Error: %v",
		e.errType.String(),
		e.code.String(),
		e.msg,
		e.err,
	)
}

// Msg returns the user-facing error message, if set.
func (e *Error) Msg() string {
	return e.msg
}

// Type returns the high-level error type.
func (e *Error) Type() Type {
	return e.errType
}

// Code returns the stable error code.
func (e *Error) Code() Code {
	return e.code
}

// Fields returns validation errors (field to message map), if any.
func (e *Error) Fields() map[string]string {
	return e.fields
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.err
}

// StatusCode maps the error code to an HTTP status code.
func (e *Error) StatusCode() int {
	switch e.code {
	case CodeInvalidFormat:
		return http.StatusBadRequest
	case CodeInvalidInput:
		return http.StatusUnprocessableEntity
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeTooManyRequest:
		return http.StatusTooManyRequests
	case CodeConflict:
		return http.StatusConflict
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func new(err error, msg string, et Type, code Code) error {
	return &Error{err: err, msg: msg, errType: et, code: code}
}

// NewServer creates a server-type error with the provided error.
func NewServer(err error) error {
	return new(err, "Internal server error", TypeServer, CodeInternal)
}

// NewBusiness creates a business-type error with the specified message and code.
func NewBusiness(msg string, code Code) error {
	return new(nil, msg, TypeBusiness, code)
}

// NewInvalidInput creates a validation error for invalid input with a message and underlying error.
func NewInvalidInput(err error, kv ...string) error {
	if err != nil {
		return new(err, "Validation error", TypeValidation, CodeInvalidInput)
	}

	if len(kv)%2 != 0 {
		return new(nil, "Invalid request body", TypeValidation, CodeInvalidFormat)
	}

	errCustomValidate := &Error{err: nil, msg: "Validation error", errType: TypeValidation, code: CodeInvalidInput}
	if errCustomValidate.fields == nil {
		errCustomValidate.fields = make(map[string]string)
	}

	for i := 0; i+1 < len(kv); i += 2 {
		errCustomValidate.fields[kv[i]] = kv[i+1]
	}

	return errCustomValidate
}

// NewInvalidFormat creates a validation error for an invalid request body format.
func NewInvalidFormat(msgs ...string) error {
	if len(msgs) == 0 {
		return new(nil, "Invalid request body", TypeValidation, CodeInvalidFormat)
	}
	return new(nil, msgs[0], TypeValidation, CodeInvalidFormat)
}
