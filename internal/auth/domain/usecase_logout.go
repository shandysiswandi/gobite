package domain

type LogoutInput struct {
	RefreshToken string `validate:"required"`
}

type LogoutOutput struct {
	Success bool
}
