package entity

const (
	UserRegistrationDestination string = "auth.user.registration"
)

type UserRegistrationMessage struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}
