package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type ChangePasswordInput struct {
	CurrentPassword string `validate:"required"`
	NewPassword     string `validate:"required,password"`
}

type ChangePasswordOutput struct {
	Success bool
}

func (s *Usecase) ChangePassword(ctx context.Context, in ChangePasswordInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	uid := clm.GetInt64(keyPayloadUserID)
	user, err := s.repoDB.GetUserCredentialInfo(ctx, uid)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", uid)
		return goerror.NewBusiness("user account not found", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "user_id", uid, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.ensureUserAllowedToLogin(ctx, user.ID, user.Status); err != nil {
		return err
	}

	if !s.password.Verify(user.Password, strings.TrimSpace(in.CurrentPassword)) {
		slog.WarnContext(ctx, "current password mismatch", "user_id", user.ID)
		return goerror.NewBusiness("invalid password", goerror.CodeUnauthorized)
	}

	// newHash, err := s.password.Hash(strings.TrimSpace(in.NewPassword))
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to hash new password", "user_id", user.ID, "error", err)
	// 	return goerror.NewServer(err)
	// }

	// if err := s.repoDB.UserCredentialUpdate(ctx, user.ID, string(newHash)); err != nil {
	// 	slog.ErrorContext(ctx, "failed to update user password", "user_id", user.ID, "error", err)
	// 	return goerror.NewServer(err)
	// }

	return nil
}
