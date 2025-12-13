package domain

type Login2FAInput struct {
	PreAuthToken string `validate:"required"`
	Code         string `validate:"required,len=6,numeric"`
}

type Login2FAOutput struct {
	AccessToken  string
	RefreshToken string
}
