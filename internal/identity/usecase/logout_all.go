package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type LogoutAllInput struct{}

func (s *Usecase) LogoutAll(ctx context.Context, in LogoutAllInput) error {
	ctx, span := s.startSpan(ctx, "LogoutAll")
	defer span.End()

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	if err := s.repoDB.RevokeAllRefreshToken(ctx, clm.UserID); err != nil {
		slog.ErrorContext(ctx, "failed to repo revoke all refresh token", "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
