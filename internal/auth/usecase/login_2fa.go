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

func (s *Usecase) Login2FA(ctx context.Context, in domain.Login2FAInput) (*domain.Login2FAOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, err
	}

	tempClaims, err := s.jwtTempToken.Verify(in.PreAuthToken)
	if err != nil {
		slog.WarnContext(ctx, "invalid pre-auth token", "error", err)
		return nil, pkgerror.ErrUnauthenticated
	}

	userID, err := strconv.ParseInt(tempClaims.Subject(), 10, 64)
	if err != nil {
		slog.WarnContext(ctx, "invalid pre-auth subject", "subject", tempClaims.Subject(), "error", err)
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

	mfaFacs, err := s.repoDB.MfaFactorGetByUserID(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get mfa factor by user_id", "user_id", user.ID, "error", err)
		return nil, err
	}

	if len(mfaFacs) == 0 {
		slog.WarnContext(ctx, "no active mfa factors for user", "user_id", user.ID)
		return nil, pkgerror.ErrUnauthenticated
	}

	payload := tempClaims.Payload()
	mfaID, ok := pickMFAID(payload["some_id"])
	if !ok {
		slog.WarnContext(ctx, "missing mfa factor in pre-auth payload", "user_id", user.ID)
		return nil, pkgerror.ErrUnauthenticated
	}

	factor, ok := findMFAFactorByID(mfaFacs, mfaID)
	if !ok {
		slog.WarnContext(ctx, "mfa factor mismatch for user", "user_id", user.ID, "mfa_id", mfaID)
		return nil, pkgerror.ErrUnauthenticated
	}

	if factor.Type != domain.MfaTypeTOTP {
		slog.WarnContext(ctx, "unsupported mfa factor type for login 2fa", "user_id", user.ID, "mfa_id", mfaID, "type", factor.Type)
		return nil, pkgerror.ErrUnauthenticated
	}

	secret := string(factor.Secret)
	if secret == "" {
		slog.WarnContext(ctx, "empty totp secret", "user_id", user.ID, "mfa_id", mfaID)
		return nil, pkgerror.ErrUnauthenticated
	}

	if !s.totp.Validate(in.Code, secret, s.clock.Now()) {
		slog.WarnContext(ctx, "invalid totp code", "user_id", user.ID, "mfa_id", mfaID)
		return nil, pkgerror.ErrUnauthenticated
	}

	subject := strconv.FormatInt(user.ID, 10)

	acToken, acJTI, err := s.jwtAccessToken.Generate(subject, pkgjwt.AccessTokenPayload{
		UserID: subject,
		Email:  user.Email,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", user.ID, "error", err)
		return nil, err
	}

	refToken, refJTI, err := s.jwtRefreshToken.Generate(subject, pkgjwt.RefreshTokenPayload{Message: "hack me"})
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate refresh jwt token", "user_id", user.ID, "error", err)
		return nil, err
	}

	if err := s.repoCache.SaveTokensID(ctx, acJTI, refJTI); err != nil {
		slog.ErrorContext(ctx, "failed to save jtis to redis", "ac", acJTI, "ref", refJTI, "error", err)
		return nil, err
	}

	return &domain.Login2FAOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}

func pickMFAID(val any) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		return id, err == nil
	default:
		return 0, false
	}
}

func findMFAFactorByID(factors []domain.MfaFactor, id int64) (*domain.MfaFactor, bool) {
	for _, f := range factors {
		if f.ID == id {
			return &f, true
		}
	}

	return nil, false
}
