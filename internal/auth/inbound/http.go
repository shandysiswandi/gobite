package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type uc interface {
	Login(ctx context.Context, in usecase.LoginInput) (*usecase.LoginOutput, error)
	LoginMFA(ctx context.Context, in usecase.LoginMFAInput) (*usecase.LoginMFAOutput, error)
	Register(ctx context.Context, in usecase.RegisterInput) error
	RegisterResend(ctx context.Context, in usecase.RegisterResendInput) error
	EmailVerify(ctx context.Context, in usecase.EmailVerifyInput) error
	ForgotPassword(ctx context.Context, in usecase.ForgotPasswordInput) error
	ResetPassword(ctx context.Context, in usecase.ResetPasswordInput) error
	Logout(ctx context.Context, in usecase.LogoutInput) error
	LogoutAll(ctx context.Context, in usecase.LogoutAllInput) error
	ChangePassword(ctx context.Context, in usecase.ChangePasswordInput) error
	RefreshToken(ctx context.Context, in usecase.RefreshTokenInput) (*usecase.RefreshTokenOutput, error)
	Profile(ctx context.Context, in usecase.ProfileInput) (*usecase.ProfileOutput, error)
	UpdateProfile(ctx context.Context, in usecase.UpdateProfileInput) error
	SetupTOTP(ctx context.Context, in usecase.SetupTOTPInput) (*usecase.SetupTOTPOutput, error)
	ConfirmTOTP(ctx context.Context, in usecase.ConfirmTOTPInput) error
}

func RegisterHTTPEndpoint(r *router.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	// Auth & User Management
	r.POST("/api/v1/auth/login", end.Login)
	r.POST("/api/v1/auth/login/mfa", end.LoginMFA)
	r.POST("/api/v1/auth/register", end.Register)
	r.POST("/api/v1/auth/register/resend", end.RegisterResend)
	r.POST("/api/v1/auth/refresh", end.RefreshToken)
	r.POST("/api/v1/auth/logout", end.Logout)        // need authenticated
	r.POST("/api/v1/auth/logout-all", end.LogoutAll) // need authenticated
	//
	r.POST("/api/v1/auth/email/verify", end.EmailVerify)

	// Password Management
	r.POST("/api/v1/auth/password/forgot", end.ForgotPassword)
	r.POST("/api/v1/auth/password/reset", end.ResetPassword)
	r.POST("/api/v1/auth/password/change", end.ChangePassword) // need authenticated

	// MFA (TOTP)
	r.POST("/api/v1/auth/mfa/totp/setup", end.SetupTOTP)     // need authenticated
	r.POST("/api/v1/auth/mfa/totp/confirm", end.ConfirmTOTP) // need authenticated

	// User Profile (need authenticated)
	r.GET("/api/v1/profile", end.Profile)
	r.PUT("/api/v1/profile", end.UpdateProfile)
}
