package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
)

type usecase interface {
	Login(ctx context.Context, in domain.LoginInput) (*domain.LoginOutput, error)
	Login2FA(ctx context.Context, in domain.Login2FAInput) (*domain.Login2FAOutput, error)
	Register(ctx context.Context, in domain.RegisterInput) (*domain.RegisterOutput, error)
	EmailVerify(ctx context.Context, in domain.EmailVerifyInput) error
	ForgotPassword(ctx context.Context, in domain.ForgotPasswordInput) (*domain.ForgotPasswordOutput, error)
	ResetPassword(ctx context.Context, in domain.ResetPasswordInput) error
	Logout(ctx context.Context, in domain.LogoutInput) error
	ChangePassword(ctx context.Context, in domain.ChangePasswordInput) error
	RefreshToken(ctx context.Context, in domain.RefreshTokenInput) (*domain.RefreshTokenOutput, error)
	Profile(ctx context.Context, in domain.ProfileInput) (*domain.ProfileOutput, error)
}

func RegisterHTTPEndpoint(r *pkgrouter.Router, uc usecase) {
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
