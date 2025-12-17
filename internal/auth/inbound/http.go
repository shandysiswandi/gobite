package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
)

type uc interface {
	Login(ctx context.Context, in usecase.LoginInput) (*usecase.LoginOutput, error)
	Login2FA(ctx context.Context, in usecase.Login2FAInput) (*usecase.Login2FAOutput, error)
	Register(ctx context.Context, in usecase.RegisterInput) (*usecase.RegisterOutput, error)
	EmailVerify(ctx context.Context, in usecase.EmailVerifyInput) error
	ForgotPassword(ctx context.Context, in usecase.ForgotPasswordInput) (*usecase.ForgotPasswordOutput, error)
	ResetPassword(ctx context.Context, in usecase.ResetPasswordInput) error
	Logout(ctx context.Context, in usecase.LogoutInput) error
	ChangePassword(ctx context.Context, in usecase.ChangePasswordInput) error
	RefreshToken(ctx context.Context, in usecase.RefreshTokenInput) (*usecase.RefreshTokenOutput, error)
	Profile(ctx context.Context, in usecase.ProfileInput) (*usecase.ProfileOutput, error)
}

func RegisterHTTPEndpoint(r *pkgrouter.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	// Auth & User Management
	r.POST("/auth/login", end.Login)
	r.POST("/auth/login-2fa", end.Login2FA)
	r.POST("/auth/register", end.Register)
	r.POST("/auth/refresh-token", end.RefreshToken)
	r.POST("/auth/logout", end.Logout) // need authenticated
	//
	r.POST("/auth/email/verify", end.EmailVerify)

	// Password Management
	r.POST("/auth/password/forgot", end.ForgotPassword)
	r.POST("/auth/password/reset", end.ResetPassword)
	r.POST("/auth/password/change", end.ChangePassword) // need authenticated

	// User Profile (need authenticated)
	r.GET("/profile", end.Profile)
}
