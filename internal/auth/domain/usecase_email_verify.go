package domain

type EmailVerifyInput struct {
	Token string `validate:"required"`
}
