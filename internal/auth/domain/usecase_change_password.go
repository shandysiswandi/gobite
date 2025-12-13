package domain

type ChangePasswordInput struct {
	CurrentPassword string `validate:"required"`
	NewPassword     string `validate:"required,password"`
}

type ChangePasswordOutput struct {
	Success bool
}
