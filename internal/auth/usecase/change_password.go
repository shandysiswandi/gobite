package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) ChangePassword(ctx context.Context, in domain.ChangePasswordInput) (*domain.ChangePasswordOutput, error) {
	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	slog.Info("a", "clm", clm)

	if err := s.validator.Validate(in); err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(clm.Subject(), 10, 64)
	if err != nil {
		slog.WarnContext(ctx, "invalid access token subject", "subject", clm.Subject(), "error", err)
		return nil, pkgerror.ErrUnauthenticated
	}

	user, err := s.repoDB.UserGetByID(ctx, userID)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", userID)
		return nil, pkgerror.ErrUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
		return nil, err
	}

	if user.Status == domain.UserStatusUnverified {
		slog.WarnContext(ctx, "user not verified", "user_id", userID)
		return nil, pkgerror.ErrAuthNotVerified
	}

	if user.Status == domain.UserStatusBanned {
		slog.WarnContext(ctx, "user account banned", "user_id", userID)
		return nil, pkgerror.ErrAuthBanned
	}

	cred, err := s.repoDB.UserCredentialGetByUserID(ctx, user.ID)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user credential not found", "user_id", user.ID)
		return nil, pkgerror.ErrUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user credential by user_id", "user_id", user.ID, "error", err)
		return nil, err
	}

	if !s.hash.Verify(cred.Password, in.CurrentPassword) {
		slog.WarnContext(ctx, "current password mismatch", "user_id", user.ID)
		return nil, pkgerror.ErrAuthUnauthenticated
	}

	newHash, err := s.hash.Hash(in.NewPassword)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash new password", "user_id", user.ID, "error", err)
		return nil, err
	}

	if err := s.repoDB.UserCredentialUpdate(ctx, user.ID, string(newHash)); err != nil {
		slog.ErrorContext(ctx, "failed to update user password", "user_id", user.ID, "error", err)
		return nil, err
	}

	return &domain.ChangePasswordOutput{Success: true}, nil
}
