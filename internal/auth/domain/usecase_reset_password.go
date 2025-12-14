package domain

type ResetPasswordInput struct {
	Token       string `validate:"required"`
	NewPassword string `validate:"required,password"`
}
