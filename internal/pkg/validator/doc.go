// Package validator provides a small validation abstraction for request and
// domain structs.
//
// Business code should depend on the Validator interface so validation can be
// shared and tested consistently. Concrete implementations (for example
// go-playground/validator v10) live in this package.
package validator
