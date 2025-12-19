package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mfacrypto"
)

type LoginMFAInput struct {
	ChallengeToken string `validate:"required"`
	Code           string `validate:"required,len=6,numeric"`
}

type LoginMFAOutput struct {
	AccessToken  string
	RefreshToken string
}

func (s *Usecase) LoginMFA(ctx context.Context, in LoginMFAInput) (*LoginMFAOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	cTokenHash, err := s.hash.Hash(in.ChallengeToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return nil, goerror.NewServer(err)
	}

	cu, err := s.repoDB.GetChallengeUserByTokenPurpose(ctx, string(cTokenHash), entity.ChallengePurposeMFALogin)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "challenge user not found", "challenge_token", string(cTokenHash))
		return nil, goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get challange user by token purpose", "challenge_token", string(cTokenHash), "error", err)
		return nil, goerror.NewServer(err)
	}

	if err := s.ensureUserAllowedToLogin(ctx, cu.UserID, cu.UserStatus); err != nil {
		return nil, err
	}

	mfaFacs, err := s.repoDB.GetMFAFactorByUserID(ctx, cu.UserID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get mfa factor by user_id", "user_id", cu.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	var factor *entity.MFAFactor
	for _, f := range mfaFacs {
		// note: only support TOTP for now
		if f.Type == entity.MFATypeTOTP {
			factor = &f
		}
	}

	if factor == nil {
		slog.WarnContext(ctx, "mfa factor for totp not found", "user_id", cu.UserID)
		return nil, goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	secretBytes, err := s.mfaCrypto.Decrypt(factor.Secret, mfacrypto.Scope{
		UserID:  cu.UserID,
		Purpose: mfacrypto.PurposeOTPSeed,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to decrypt totp secret", "user_id", cu.UserID, "mfa_id", factor.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if !s.totp.Validate(in.Code, string(secretBytes), s.clock.Now()) {
		slog.WarnContext(ctx, "invalid totp code", "user_id", cu.UserID, "mfa_id", factor.ID)
		return nil, goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	if err := s.repoDB.DeleteChallenge(ctx, cu.ChallengeID); err != nil {
		slog.ErrorContext(ctx, "failed to repo delete challenge by id", "challenge_id", cu.ChallengeID, "error", err)
		return nil, goerror.NewServer(err)
	}

	acJTI := s.uuid.Generate()
	acToken, err := s.jwt.Generate(
		jwt.WithID(acJTI),
		jwt.WithSubject(strconv.FormatInt(cu.UserID, 10)),
		jwt.WithPayloadValue(keyPayloadUserID, cu.UserID),
		jwt.WithPayloadValue(keyPayloadUserEmail, cu.UserEmail),
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", cu.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	refToken := s.oid.Generate()
	refTokenHash, err := s.hash.Hash(refToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash refresh token", "user_id", cu.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if err := s.repoDB.CreateRefreshToken(ctx, entity.RefreshToken{
		ID:        s.uid.Generate(),
		UserID:    cu.UserID,
		Token:     string(refTokenHash),
		ExpiresAt: s.clock.Now().Add(time.Duration(s.cfg.GetInt("modules.auth.refresh_token_ttl")) * 24 * time.Hour),
	}); err != nil {
		slog.ErrorContext(ctx, "failed to repo create refresh token user", "user_id", cu.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &LoginMFAOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
