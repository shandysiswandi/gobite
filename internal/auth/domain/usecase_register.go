package domain

type RegisterInput struct {
	Email    string `validate:"required,lowercase,email"`
	Password string `validate:"required,password"`
	FullName string `validate:"required,min=2,max=100,alphaspace"`
}

type RegisterOutput struct {
	IsNeedVerify bool
}
