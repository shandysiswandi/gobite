package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type ResetPasswordInput struct {
	Token       string `validate:"required"`
	NewPassword string `validate:"required,password"`
}

func (s *Usecase) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	// cTokenHash, err := s.hash.Hash(in.Token)
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to hash token", "error", err)
	// 	return goerror.NewServer(err)
	// }

	// now := s.clock.Now()
	// err = s.repoDB.UserPasswordResetConsume(ctx, in.Token, string(newHash), now)
	// if errors.Is(err, goerror.ErrNotFound) {
	// 	return goerror.NewBusiness("invalid or expired reset token", goerror.CodeUnauthorized)
	// }
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to consume password reset token", "error", err)
	// 	return goerror.NewServer(err)
	// }

	return nil
}
