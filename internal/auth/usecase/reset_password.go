package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

type ResetPasswordInput struct {
	Token       string `validate:"required"`
	NewPassword string `validate:"required,password"`
}

func (s *Usecase) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	newHash, err := s.hash.Hash(in.NewPassword)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash new password", "error", err)
		return pkgerror.NewServer(err)
	}

	now := s.clock.Now()
	if err := s.repoDB.UserPasswordResetConsume(ctx, in.Token, string(newHash), now); err != nil {
		if errors.Is(err, pkgerror.ErrNotFound) {
			return pkgerror.NewBusiness("invalid or expired reset token", pkgerror.CodeUnauthorized)
		}

		slog.ErrorContext(ctx, "failed to consume password reset token", "error", err)
		return pkgerror.NewServer(err)
	}

	return nil
}
