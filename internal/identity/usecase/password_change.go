package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type PasswordChangeInput struct {
	CurrentPassword string `validate:"required"`
	NewPassword     string `validate:"required,password"`
}

func (s *Usecase) PasswordChange(ctx context.Context, in PasswordChangeInput) error {
	ctx, span := s.startSpan(ctx, "PasswordChange")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	user, err := s.repoDB.GetUserCredentialInfo(ctx, clm.UserID)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", clm.UserID)
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user credential info", "user_id", clm.UserID, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.ensureUserStatusAllowed(ctx, user.ID, user.Status); err != nil {
		return err
	}

	if !s.bcrypt.Verify(user.Password, in.CurrentPassword) {
		slog.WarnContext(ctx, "current password mismatch", "user_id", user.ID)
		return goerror.NewBusiness("invalid password", goerror.CodeUnauthorized)
	}

	newHash, err := s.bcrypt.Hash(in.NewPassword)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash new password", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.repoDB.UpdateUserCredential(ctx, user.ID, string(newHash)); err != nil {
		slog.ErrorContext(ctx, "failed to update user password", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
