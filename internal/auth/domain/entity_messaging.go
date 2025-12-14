package domain

const (
	UserRegistrationDestination string = "auth_user_registration"
)

type UserRegistrationMessage struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}
