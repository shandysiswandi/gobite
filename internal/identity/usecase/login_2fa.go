package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/mfa"
)

type Login2FAInput struct {
	ChallengeToken string         `validate:"required"`
	Method         entity.MFAType `validate:"required"`
	Code           string         `validate:"required"`
}

type Login2FAOutput struct {
	AccessToken  string
	RefreshToken string
}

func (s *Usecase) Login2FA(ctx context.Context, in Login2FAInput) (*Login2FAOutput, error) {
	ctx, span := s.startSpan(ctx, "Login2FA")
	defer span.End()

	in.Code = strings.TrimSpace(in.Code)

	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	if in.Method == entity.MFATypeUnknown || in.Method == entity.MFATypeSMS {
		slog.WarnContext(ctx, "method not supported", "method", in.Method.String())
		return nil, goerror.NewBusiness("method not supported", goerror.CodeUnauthorized)
	}

	if in.Method == entity.MFATypeTOTP && !s.isValidTOTPCode(in.Code) {
		slog.WarnContext(ctx, "totp code is not valid", "code", in.Code)
		return nil, goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	cu, err := s.loadChallengeUser(ctx, in.ChallengeToken)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUserStatusAllowed(ctx, cu.UserID, cu.UserStatus); err != nil {
		return nil, err
	}

	mfaFacs, err := s.loadVerifiedFactors(ctx, cu.UserID)
	if err != nil {
		return nil, err
	}

	if in.Method == entity.MFATypeTOTP {
		if err := s.verifyTOTP(ctx, cu.UserID, mfaFacs, in.Code); err != nil {
			return nil, err
		}
	}

	if in.Method == entity.MFATypeBackupCode {
		if err := s.verifyBackupCode(ctx, cu.UserID, mfaFacs, in.Code); err != nil {
			return nil, err
		}
	}

	return s.issueLoginTokens(ctx, cu)
}

func (s *Usecase) isValidTOTPCode(code string) bool {
	if len(code) != 6 { // 6 is length of totp
		return false
	}

	for i := 0; i < len(code); i++ {
		if code[i] < '0' || code[i] > '9' {
			return false
		}
	}

	return true
}

func (s *Usecase) loadChallengeUser(ctx context.Context, token string) (*entity.ChallengeUser, error) {
	cTokenHash, err := s.hmac.Hash(token)
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

	return cu, nil
}

func (s *Usecase) loadVerifiedFactors(ctx context.Context, userID int64) ([]entity.MFAFactor, error) {
	mfaFacs, err := s.repoDB.GetMFAFactorByUserID(ctx, userID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get mfa factor by user_id", "user_id", userID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if len(mfaFacs) == 0 {
		slog.WarnContext(ctx, "mfa factor not found for backup code", "user_id", userID)
		return nil, goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	return mfaFacs, nil
}

func (s *Usecase) verifyTOTP(ctx context.Context, userID int64, factors []entity.MFAFactor, code string) error {
	var factor *entity.MFAFactor
	for i := range factors {
		if factors[i].Type == entity.MFATypeTOTP {
			factor = &factors[i]
			break
		}
	}

	if factor == nil {
		slog.WarnContext(ctx, "mfa factor for totp not found", "user_id", userID)
		return goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	secretBytes, err := s.mfaEncryptor.Decrypt(factor.Secret, mfa.Scope{
		UserID:  userID,
		Purpose: mfa.PurposeOTPSeed,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to decrypt totp secret", "user_id", userID, "mfa_id", factor.ID, "error", err)
		return goerror.NewServer(err)
	}

	if !s.totp.Validate(code, string(secretBytes), s.clock.Now()) {
		slog.WarnContext(ctx, "invalid totp code", "user_id", userID, "mfa_id", factor.ID)
		return goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	if err := s.repoDB.UpdateMFALastUsedAt(ctx, factor.ID, userID); err != nil {
		slog.ErrorContext(ctx, "failed to update mfa last_used_at", "user_id", userID, "mfa_id", factor.ID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}

func (s *Usecase) verifyBackupCode(ctx context.Context, userID int64, factors []entity.MFAFactor, code string) error {
	var factor *entity.MFAFactor
	for i := range factors {
		if factors[i].Type == entity.MFATypeBackupCode {
			factor = &factors[i]
			break
		}
	}

	if factor == nil {
		slog.WarnContext(ctx, "mfa factor for backup code not found", "user_id", userID)
		return goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	codes, err := s.repoDB.GetMFABackupCodeByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get backup code by user id", "user_id", userID, "error", err)
		return goerror.NewServer(err)
	}

	var bc *entity.MFABackupCode
	for _, stored := range codes {
		if s.argon2id.Verify(stored.Code, code) {
			bc = &stored
			break
		}
	}

	if bc == nil {
		slog.WarnContext(ctx, "backup code not match", "user_id", userID)
		return goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	isValid, err := s.repoDB.MarkMFABackupCodeUsed(ctx, bc.ID, bc.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to consume backup code", "user_id", userID, "error", err)
		return goerror.NewServer(err)
	}
	if !isValid {
		slog.WarnContext(ctx, "backup code already used", "user_id", userID)
		return goerror.NewBusiness("invalid challenge session or code", goerror.CodeUnauthorized)
	}

	if err := s.repoDB.UpdateMFALastUsedAt(ctx, factor.ID, userID); err != nil {
		slog.ErrorContext(ctx, "failed to update mfa last_used_at", "user_id", userID, "mfa_id", factor.ID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}

func (s *Usecase) issueLoginTokens(ctx context.Context, cu *entity.ChallengeUser) (*Login2FAOutput, error) {
	acToken, err := s.jwt.Generate(cu.UserID, cu.UserEmail)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", cu.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	refToken := s.oid.Generate()
	refTokenHash, err := s.hmac.Hash(refToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash refresh token", "user_id", cu.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	refresh := entity.RefreshToken{
		ID:        s.uid.Generate(),
		UserID:    cu.UserID,
		Token:     string(refTokenHash),
		ExpiresAt: s.clock.Now().Add(s.cfg.GetDay("modules.identity.refresh_token_ttl_days")),
	}

	if err := s.repoDB.NewRefreshToken(ctx, refresh, cu.ChallengeID); err != nil {
		slog.ErrorContext(ctx, "failed to repo new refresh token user", "user_id", cu.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &Login2FAOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
