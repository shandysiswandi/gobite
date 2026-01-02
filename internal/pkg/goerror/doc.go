// Package goerror defines shared error types and sentinel errors used across
// the application.
//
// It helps keep error handling consistent by:
//   - Providing sentinel errors that can be checked with errors.Is.
//   - Providing a structured Error type that carries a message, type, and code,
//     which can be mapped to HTTP status codes at the edge (handlers).
package goerror
