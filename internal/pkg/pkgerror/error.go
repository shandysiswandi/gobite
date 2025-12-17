package pkgerror

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrNotFound indicates that the requested resource could not be found.
	ErrNotFound = errors.New("resource not found")
)

// Type classifies errors into high-level buckets used by the application.
type Type int

const (
	TypeServer     Type = iota // Server-side errors (e.g., database or network issues).
	TypeBusiness               // Business logic errors (e.g., domain rule violations).
	TypeValidation             // Validation errors (e.g., input validation failures).
)

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
	CodeInternal      Code = iota // Internal or unspecified error.
	CodeInvalidFormat             // Error code for invalid format.
	CodeInvalidInput              // Error code for invalid input.
	CodeNotFound                  // Error code for resource not found.
	CodeConflict                  // Error code for conflict situations (e.g., duplicate entries).
	CodeUnauthorized              // Error code for unauthorized access.
	CodeForbidden                 // Error code for forbidden actions.
	CodeTimeout                   // Error code for operation timeout.
)

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
func NewInvalidInput(err error) error {
	return new(err, "validation error", TypeValidation, CodeInvalidInput)
}

// NewInvalidFormat creates a validation error for an invalid request body format.
func NewInvalidFormat() error {
	return new(nil, "invalid request body", TypeValidation, CodeInvalidFormat)
}
