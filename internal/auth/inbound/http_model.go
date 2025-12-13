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

type RegisterResponse struct {
	IsNeedVerify bool `json:"is_need_verify"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutResponse struct {
	Success bool `json:"success"`
}
