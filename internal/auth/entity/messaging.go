package entity

const (
	UserRegistrationDestination   string = "auth.user.registration"
	UserForgotPasswordDestination string = "auth.user.forgot.password"
)

type UserRegistrationMessage struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type UserForgotPasswordMessage struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Token    string `json:"token"`
}
