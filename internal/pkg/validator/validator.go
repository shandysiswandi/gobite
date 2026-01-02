package validator

// Validator is an interface that defines a single method for validating data.
type Validator interface {
	// Validate validates the given data and returns an error if validation fails.
	// The input can be any type, typically a struct.
	Validate(data any) error
}
