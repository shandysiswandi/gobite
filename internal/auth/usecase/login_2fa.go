package usecase

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

func (s *Usecase) Login2FA(ctx context.Context, in domain.Login2FAInput) (*domain.Login2FAOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, pkgerror.NewInvalidInput(err)
	}

	tempClaims, err := s.jwtTempToken.Verify(in.PreAuthToken)
	if err != nil {
		slog.WarnContext(ctx, "invalid pre-auth token", "error", err)
		return nil, pkgerror.NewBusiness("invalid pre auth token", pkgerror.CodeUnauthorized)
	}

	userID, err := strconv.ParseInt(tempClaims.Subject(), 10, 64)
	if err != nil {
		slog.WarnContext(ctx, "invalid pre-auth subject", "subject", tempClaims.Subject(), "error", err)
		return nil, pkgerror.NewBusiness("invalid pre auth token", pkgerror.CodeUnauthorized)
	}

	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	mfaFacs, err := s.repoDB.MfaFactorGetByUserID(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get mfa factor by user_id", "user_id", user.ID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	if len(mfaFacs) == 0 {
		slog.WarnContext(ctx, "no active mfa factors for user", "user_id", user.ID)
		return nil, pkgerror.NewBusiness("invalid pre auth token", pkgerror.CodeUnauthorized)
	}

	payload := tempClaims.Payload()
	mfaID, ok := pickMFAID(payload["some_id"])
	if !ok {
		slog.WarnContext(ctx, "missing mfa factor in pre-auth payload", "user_id", user.ID)
		return nil, pkgerror.NewBusiness("invalid pre auth token", pkgerror.CodeUnauthorized)
	}

	factor, ok := findMFAFactorByID(mfaFacs, mfaID)
	if !ok {
		slog.WarnContext(ctx, "mfa factor mismatch for user", "user_id", user.ID, "mfa_id", mfaID)
		return nil, pkgerror.NewBusiness("invalid pre auth token", pkgerror.CodeUnauthorized)
	}

	if factor.Type != domain.MfaTypeTOTP {
		slog.WarnContext(ctx, "unsupported mfa factor type for login 2fa", "user_id", user.ID, "mfa_id", mfaID, "type", factor.Type)
		return nil, pkgerror.NewBusiness("invalid pre auth token", pkgerror.CodeUnauthorized)
	}

	secret := string(factor.Secret)
	if secret == "" {
		slog.WarnContext(ctx, "empty totp secret", "user_id", user.ID, "mfa_id", mfaID)
		return nil, pkgerror.NewServer(pkgerror.ErrNotFound) // actually this will not happend by system
	}

	if !s.totp.Validate(in.Code, secret, s.clock.Now()) {
		slog.WarnContext(ctx, "invalid totp code", "user_id", user.ID, "mfa_id", mfaID)
		return nil, pkgerror.NewBusiness("invalid code", pkgerror.CodeUnauthorized)
	}

	acToken, acJTI, refToken, refJTI, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := s.repoCache.SaveTokensID(ctx, acJTI, refJTI); err != nil {
		slog.ErrorContext(ctx, "failed to save jtis to redis", "ac", acJTI, "ref", refJTI, "error", err)
		return nil, pkgerror.NewServer(err)
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
