package domain

type RefreshTokenInput struct {
	RefreshToken string `validate:"required"`
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
}
