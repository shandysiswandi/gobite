package domain

type LoginInput struct {
	Email    string `validate:"required,lowercase,email"`
	Password string `validate:"required"`
}

type LoginOutput struct {
	MfaRequired  bool
	PreAuthToken string
	//
	AccessToken  string
	RefreshToken string
}
