package inbound

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	MfaRequired      bool     `json:"mfa_required,omitempty"`
	ChallengeToken   string   `json:"challenge_token,omitempty"`
	AvailableMethods []string `json:"available_methods,omitempty"`
	AccessToken      string   `json:"access_token,omitempty"`
	RefreshToken     string   `json:"refresh_token,omitempty"`
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

type RegisterResendResponse struct{}

func (RegisterResendResponse) Message() string {
	return "If an account with that email exists, we have sent a verification link."
}

type EmailVerifyRequest struct {
	ChallengeToken string `json:"challenge_token"`
}

type PasswordForgotRequest struct {
	Email string `json:"email"`
}

type PasswordForgotResponse struct{}

func (PasswordForgotResponse) Message() string {
	return "If an account with that email exists, we have sent a password reset link."
}

type PasswordResetRequest struct {
	ChallengeToken string `json:"challenge_token"`
	NewPassword    string `json:"new_password"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type Login2FARequest struct {
	ChallengeToken string `json:"challenge_token"`
	Method         string `json:"method"`
	Code           string `json:"code"`
}

type Login2FAResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TOTPSetupRequest struct {
	FriendlyName    string `json:"friendly_name"`
	CurrentPassword string `json:"current_password"`
}

type TOTPSetupResponse struct {
	ChallengeToken string `json:"challenge_token"`
	Key            string `json:"key"`
	URI            string `json:"uri"`
}

type TOTPConfirmRequest struct {
	ChallengeToken string `json:"challenge_token"`
	Code           string `json:"code"`
}

type BackupCodeRequest struct {
	CurrentPassword string `json:"current_password"`
}

type BackupCodeResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
}

type ProfilePermissionsResponse struct {
	Permissions map[string][]string `json:"permissions"`
}

type ProfileSettingMFAResponse struct {
	TOTPEnabled       bool `json:"totp_enabled"`
	BackupCodeEnabled bool `json:"backup_code_enabled"`
	SMSEnabled        bool `json:"sms_enabled"`
}

type ProfileResponse struct {
	ID        int64  `json:"id,string"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	Status    string `json:"status"`
}

type UserResponse struct {
	ID        int64             `json:"id,string"`
	Email     string            `json:"email"`
	FullName  string            `json:"full_name"`
	AvatarURL string            `json:"avatar_url"`
	Status    entity.UserStatus `json:"status"`
	UpdateAt  time.Time         `json:"updated_at"`
}

type UserCreateRequest struct {
	Email    string            `json:"email"`
	Password string            `json:"password"`
	FullName string            `json:"full_name"`
	Status   entity.UserStatus `json:"status"`
}

type UserUpdateRequest struct {
	Email    string            `json:"email,omitempty"`
	Password string            `json:"password,omitempty"`
	FullName string            `json:"full_name,omitempty"`
	Status   entity.UserStatus `json:"status,omitempty"`
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
	// meta
	total int64
	size  int32
	page  int32
}

func (r UsersResponse) Meta() map[string]any {
	return map[string]any{
		"total": r.total,
		"size":  r.size,
		"page":  r.page,
	}
}

type UserDetailResponse struct {
	User UserResponse `json:"user"`
}

type UserExportResponse struct {
	Users []UserResponse `json:"users"`
}

type UserImportRequest []UserImportUserRequest

type UserImportUserRequest struct {
	Email    string            `json:"email"`
	Password string            `json:"password"`
	FullName string            `json:"full_name"`
	Status   entity.UserStatus `json:"status"`
}

type UserImportResponse struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
}
