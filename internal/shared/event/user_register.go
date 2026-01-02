package event

const UserRegistrationDestination string = "user_registration"
const UserRegistrationDestinationConsumerNotification string = "user_registration_notification"

type UserRegistrationMessage struct {
	UserID         int64  `json:"user_id"`
	Email          string `json:"email"`
	FullName       string `json:"full_name"`
	ChallengeToken string `json:"challenge_token"`
}
