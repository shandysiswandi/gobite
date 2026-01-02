package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/identity/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type uc interface {
	Login(ctx context.Context, in usecase.LoginInput) (*usecase.LoginOutput, error)
	Login2FA(ctx context.Context, in usecase.Login2FAInput) (*usecase.Login2FAOutput, error)
	RefreshToken(ctx context.Context, in usecase.RefreshTokenInput) (*usecase.RefreshTokenOutput, error)

	Register(ctx context.Context, in usecase.RegisterInput) error
	RegisterResend(ctx context.Context, in usecase.RegisterResendInput) error
	RegisterVerify(ctx context.Context, in usecase.RegisterVerifyInput) error

	PasswordForgot(ctx context.Context, in usecase.PasswordForgotInput) error
	PasswordReset(ctx context.Context, in usecase.PasswordResetInput) error
	PasswordChange(ctx context.Context, in usecase.PasswordChangeInput) error

	Logout(ctx context.Context, in usecase.LogoutInput) error
	LogoutAll(ctx context.Context, in usecase.LogoutAllInput) error

	Profile(ctx context.Context, in usecase.ProfileInput) (*usecase.ProfileOutput, error)
	ProfileUpdate(ctx context.Context, in usecase.ProfileUpdateInput) error
	ProfileUpdateAvatar(ctx context.Context, in usecase.ProfileUpdateAvatarInput) error

	TOTPSetup(ctx context.Context, in usecase.TOTPSetupInput) (*usecase.TOTPSetupOutput, error)
	TOTPConfirm(ctx context.Context, in usecase.TOTPConfirmInput) (*usecase.TOTPConfirmOutput, error)
	BackupCodeRotate(ctx context.Context, in usecase.BackupCodeRotateInput) (*usecase.BackupCodeRotateOutput, error)
}

func RegisterHTTPEndpoint(r *router.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	// Auth & User Management
	r.POST("/api/v1/identity/login", end.Login)
	r.POST("/api/v1/identity/login/2fa", end.Login2FA)
	r.POST("/api/v1/identity/refresh", end.RefreshToken)
	//
	r.POST("/api/v1/identity/register", end.Register)
	r.POST("/api/v1/identity/register/resend", end.RegisterResend)
	r.POST("/api/v1/identity/register/verify", end.RegisterVerify)
	//
	r.POST("/api/v1/identity/logout", end.Logout)
	r.POST("/api/v1/identity/logout-all", end.LogoutAll) // need authenticated

	// Password Management
	r.POST("/api/v1/identity/password/forgot", end.PasswordForgot)
	r.POST("/api/v1/identity/password/reset", end.PasswordReset)
	r.POST("/api/v1/identity/password/change", end.PasswordChange) // need authenticated

	// MFA (TOTP)
	r.POST("/api/v1/identity/mfa/totp/setup", end.TOTPSetup)                // need authenticated
	r.POST("/api/v1/identity/mfa/totp/confirm", end.TOTPConfirm)            // need authenticated
	r.POST("/api/v1/identity/mfa/backup_code/rotate", end.BackupCodeRotate) // need authenticated

	// User Profile (need authenticated)
	r.GET("/api/v1/identity/profile", end.Profile)
	r.PUT("/api/v1/identity/profile", end.ProfileUpdate)
	r.PUT("/api/v1/identity/profile/avatar", end.ProfileUpdateAvatar)
}
