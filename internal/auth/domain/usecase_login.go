package domain

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}
