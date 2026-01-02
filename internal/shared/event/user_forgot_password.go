package event

const UserForgotPasswordDestination string = "user_forgot_password"
const UserForgotPasswordConsumerNotification string = "user_forgot_password_notification"

type UserForgotPasswordMessage struct {
	UserID         int64  `json:"user_id"`
	Email          string `json:"email"`
	ChallengeToken string `json:"challenge_token"`
}
