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
		r.Post("/register", end.Register)
		r.Post("/logout", end.Logout)
		r.Post("/refresh-token", end.RefreshToken)
	})
}

type usecase interface {
	Login(ctx context.Context, in domain.LoginInput) (*domain.LoginOutput, error)
	Register(ctx context.Context, in domain.RegisterInput) (*domain.RegisterOutput, error)
	Logout(ctx context.Context, in domain.LogoutInput) (*domain.LogoutOutput, error)
	RefreshToken(ctx context.Context, in domain.RefreshTokenInput) (*domain.RefreshTokenOutput, error)
}
