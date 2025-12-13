package domain

type ProfileInput struct{}

type ProfileOutput struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    string
}
