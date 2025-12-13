package inbound

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/shandysiswandi/gobite/internal/auth/domain"
)

type HTTPEndpoint struct {
	uc usecase
}

func RegisterHTTPEndpoint(r chi.Router, uc usecase) {
	end := &HTTPEndpoint{uc: uc}

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", end.Login)
		r.Post("/login-2fa", end.Login2FA)
		r.Post("/register", end.Register)
		r.Post("/logout", end.Logout)
		r.Post("/refresh-token", end.RefreshToken)
	})

	r.Route("/me", func(r chi.Router) {
		r.Put("/change-password", end.ChangePassword)
		r.Get("/profile", end.Profile)
	})
}

type usecase interface {
	Login(ctx context.Context, in domain.LoginInput) (*domain.LoginOutput, error)
	Login2FA(ctx context.Context, in domain.Login2FAInput) (*domain.Login2FAOutput, error)
	Register(ctx context.Context, in domain.RegisterInput) (*domain.RegisterOutput, error)
	Logout(ctx context.Context, in domain.LogoutInput) (*domain.LogoutOutput, error)
	ChangePassword(ctx context.Context, in domain.ChangePasswordInput) (*domain.ChangePasswordOutput, error)
	RefreshToken(ctx context.Context, in domain.RefreshTokenInput) (*domain.RefreshTokenOutput, error)
	Profile(ctx context.Context, in domain.ProfileInput) (*domain.ProfileOutput, error)
}
