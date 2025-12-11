package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

func (s *Usecase) Login(ctx context.Context, in domain.LoginInput) (*domain.LoginOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, err
	}

	user, err := s.repoDB.UserGetByEmail(ctx, in.Email)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found")
		return nil, pkgerror.ErrAuthUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "error", err)
		return nil, err
	}

	if user.Status == domain.UserStatusUnverified {
		slog.WarnContext(ctx, "user not verified")
		return nil, pkgerror.ErrAuthNotVerified
	}

	if user.Status == domain.UserStatusBanned {
		slog.WarnContext(ctx, "user account banned")
		return nil, pkgerror.ErrAuthBanned
	}

	userCred, err := s.repoDB.UserCredentialGetByUserID(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user credential by user_id", "user_id", user.ID, "error", err)
		return nil, err
	}

	if !s.hash.Verify(userCred.Password, in.Password) {
		slog.WarnContext(ctx, "password user account not match")
		return nil, pkgerror.ErrAuthUnauthenticated
	}

	mfaFacs, err := s.repoDB.MfaFactorGetByUserID(ctx, user.ID)

	return &domain.LoginOutput{
		AccessToken:  "AccessToken",
		RefreshToken: "AccessToken",
	}, nil
}
