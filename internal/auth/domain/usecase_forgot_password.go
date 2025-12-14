package domain

type ForgotPasswordInput struct {
	Email string `validate:"required,lowercase,email"`
}

type ForgotPasswordOutput struct {
	Success bool
}
