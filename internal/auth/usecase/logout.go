package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type LogoutInput struct {
	RefreshToken string `validate:"required"`
}

func (s *Usecase) Logout(ctx context.Context, in LogoutInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	tokenHash, err := s.hash.Hash(in.RefreshToken)
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
