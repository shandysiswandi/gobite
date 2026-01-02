package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type PasswordResetInput struct {
	ChallengeToken string `validate:"required"`
	NewPassword    string `validate:"required,password"`
}

func (s *Usecase) PasswordReset(ctx context.Context, in PasswordResetInput) error {
	ctx, span := s.startSpan(ctx, "PasswordReset")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	cTokenHash, err := s.hmac.Hash(in.ChallengeToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token", "error", err)
		return goerror.NewServer(err)
	}

	cu, err := s.repoDB.GetChallengeUserByTokenPurpose(ctx, string(cTokenHash), entity.ChallengePurposePasswordForgotReset)
	if errors.Is(err, goerror.ErrNotFound) {
		return goerror.NewBusiness("invalid or expired reset token", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get challenge user by token purpose", "challenge_token", string(cTokenHash), "error", err)
		return goerror.NewServer(err)
	}

	if err := s.ensureUserStatusAllowed(ctx, cu.UserID, cu.UserStatus); err != nil {
		return err
	}

	newHash, err := s.bcrypt.Hash(in.NewPassword)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash new password", "user_id", cu.UserID, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.repoDB.ResetUserPassword(ctx, cu.UserID, cu.ChallengeID, string(newHash)); err != nil {
		slog.ErrorContext(ctx, "failed to update user password", "user_id", cu.UserID, "challenge_id", cu.ChallengeID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
