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
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user credential not found", "user_id", user.ID)
		return nil, pkgerror.ErrAuthUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user credential by user_id", "user_id", user.ID, "error", err)
		return nil, err
	}

	if !s.hash.Verify(userCred.Password, in.Password) {
		slog.WarnContext(ctx, "password user account not match")
		return nil, pkgerror.ErrAuthUnauthenticated
	}

	mfaFacs, err := s.repoDB.MfaFactorGetByUserID(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get mfa factor by user_id", "user_id", user.ID, "error", err)
		return nil, err
	}

	strID := strconv.FormatInt(user.ID, 10)

	// this mean user has mfa active
	if len(mfaFacs) > 0 {
		tempToken, _, err := s.jwtTempToken.Generate(strID, map[string]any{"some_id": mfaFacs[0].ID})
		if err != nil {
			slog.ErrorContext(ctx, "failed to generate temp jwt token", "user_id", user.ID, "error", err)
			return nil, err
		}

		return &domain.LoginOutput{
			MfaRequired:  true,
			PreAuthToken: tempToken,
		}, nil
	}

	acToken, acJTI, err := s.jwtAccessToken.Generate(strID, pkgjwt.AccessTokenPayload{Email: user.Email})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", user.ID, "error", err)
		return nil, err
	}

	refToken, refJTI, err := s.jwtRefreshToken.Generate(strID, pkgjwt.RefreshTokenPayload{Message: "hack me"})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate refresh jwt token", "user_id", user.ID, "error", err)
		return nil, err
	}

	if err := s.repoCache.SaveTokensID(ctx, acJTI, refJTI); err != nil {
		slog.ErrorContext(ctx, "failed to save jtis to redis", "ac", acJTI, "ref", refJTI, "error", err)
		return nil, err
	}

	return &domain.LoginOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
