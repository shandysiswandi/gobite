package inbound

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginResponse struct {
	MfaRequired  bool   `json:"mfa_required,omitempty"`
	PreAuthToken string `json:"pre_auth_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type Login2FARequest struct {
	PreAuthToken string `json:"pre_auth_token"`
	Code         string `json:"code"`
}

type Login2FAResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterResponse struct {
	IsNeedVerify bool `json:"is_need_verify"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutResponse struct {
	Success bool `json:"success"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ChangePasswordResponse struct {
	Success bool `json:"success"`
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
