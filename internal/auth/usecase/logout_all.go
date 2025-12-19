package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type LogoutAllInput struct{}

func (s *Usecase) LogoutAll(ctx context.Context, in LogoutAllInput) error {
	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	userID := clm.GetInt64(keyPayloadUserID)
	if err := s.repoDB.RevokeAllRefreshToken(ctx, userID); err != nil {
		slog.ErrorContext(ctx, "failed to repo revoke all refresh token", "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
