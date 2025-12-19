package inbound

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	MfaRequired    bool   `json:"mfa_required,omitempty"`
	ChallengeToken string `json:"challenge_token,omitempty"`
	AccessToken    string `json:"access_token,omitempty"`
	RefreshToken   string `json:"refresh_token,omitempty"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type RegisterResponse struct{}

func (RegisterResponse) Message() string {
	return "Registration successful. Please check your email to verify your account."
}

type RegisterResendRequest struct {
	Email string `json:"email"`
}

type RegisterResendesponse struct{}

func (RegisterResendesponse) Message() string {
	return "If an account with that email exists, we have sent a verification link."
}

type EmailVerifyRequest struct {
	Token string `json:"token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type LoginMFARequest struct {
	ChallengeToken string `json:"challenge_token"`
	Code           string `json:"code"`
}

type LoginMFAResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type SetupTOTPRequest struct {
	FriendlyName    string `json:"friendly_name"`
	CurrentPassword string `json:"current_password"`
}

type SetupTOTPResponse struct {
	ChallengeToken string `json:"challenge_token"`
	Key            string `json:"key"`
	URI            string `json:"uri"`
}

type ConfirmTOTPRequest struct {
	ChallengeToken string `json:"challenge_token"`
	Code           string `json:"code"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
}

type ProfileResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	Status    string `json:"status"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
