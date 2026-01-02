package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type LogoutInput struct {
	RefreshToken string
}

func (s *Usecase) Logout(ctx context.Context, in LogoutInput) error {
	ctx, span := s.startSpan(ctx, "Logout")
	defer span.End()

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	if len(in.RefreshToken) != 64 {
		return nil
	}

	tokenHash, err := s.hmac.Hash(in.RefreshToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash refresh token", "error", err)
		return goerror.NewServer(err)
	}

	if err := s.repoDB.RevokeRefreshToken(ctx, string(tokenHash)); err != nil {
		slog.ErrorContext(ctx, "failed to repo revoke refresh token", "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
