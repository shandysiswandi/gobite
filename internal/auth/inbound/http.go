package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
)

type HTTPEndpoint struct {
	uc usecase
}

func RegisterHTTPEndpoint(r *pkgrouter.Router, uc usecase) {
	end := &HTTPEndpoint{uc: uc}

	r.POST("/auth/login", end.Login)
	r.POST("/auth/login-2fa", end.Login2FA)
	r.POST("/auth/register", end.Register)
	r.POST("/auth/forgot-password", end.ForgotPassword)
	r.POST("/auth/logout", end.Logout)
	r.POST("/auth/refresh-token", end.RefreshToken)

	r.PUT("/me/change-password", end.ChangePassword)
	r.GET("/me/profile", end.Profile)
}

type usecase interface {
	Login(ctx context.Context, in domain.LoginInput) (*domain.LoginOutput, error)
	Login2FA(ctx context.Context, in domain.Login2FAInput) (*domain.Login2FAOutput, error)
	Register(ctx context.Context, in domain.RegisterInput) (*domain.RegisterOutput, error)
	ForgotPassword(ctx context.Context, in domain.ForgotPasswordInput) (*domain.ForgotPasswordOutput, error)
	Logout(ctx context.Context, in domain.LogoutInput) error
	ChangePassword(ctx context.Context, in domain.ChangePasswordInput) error
	RefreshToken(ctx context.Context, in domain.RefreshTokenInput) (*domain.RefreshTokenOutput, error)
	Profile(ctx context.Context, in domain.ProfileInput) (*domain.ProfileOutput, error)
}
