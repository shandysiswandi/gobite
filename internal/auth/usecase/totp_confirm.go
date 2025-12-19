package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type ConfirmTOTPInput struct {
	ChallengeToken string `validate:"required"`
	Code           string `validate:"required,len=6,numeric"`
}

func (s *Usecase) ConfirmTOTP(ctx context.Context, in ConfirmTOTPInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	cTokenHash, err := s.hash.Hash(in.ChallengeToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return goerror.NewServer(err)
	}

	cu, err := s.repoDB.GetChallengeUserByTokenPurpose(ctx, string(cTokenHash), entity.ChallengePurposeMFASetupConfirm)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "challenge user not found", "challenge_token", string(cTokenHash))
		return goerror.NewBusiness("invalid challenge session", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get challange user by token purpose", "challenge_token", string(cTokenHash), "error", err)
		return goerror.NewServer(err)
	}

	if err := s.ensureUserAllowedToLogin(ctx, cu.UserID, cu.UserStatus); err != nil {
		return err
	}

	// mfaID := cu.ChallengeMetadata.GetInt64("mfa_id")
	// factor, err := s.repoDB.MfaFactorGetByID(ctx, mfaID, cu.UserID)
	// if errors.Is(err, goerror.ErrNotFound) {
	// 	slog.WarnContext(ctx, "mfa factor not found", "user_id", cu.UserID, "mfa_id", mfaID)
	// 	return goerror.NewBusiness("mfa factor totp not found", goerror.CodeNotFound)
	// }
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to repo get mfa factor", "user_id", cu.UserID, "mfa_id", mfaID, "error", err)
	// 	return goerror.NewServer(err)
	// }

	// if factor.Type != entity.MFATypeTOTP {
	// 	slog.WarnContext(ctx, "unsupported mfa factor type for verification", "user_id", cu.UserID, "mfa_id", mfaID, "type", factor.Type.String())
	// 	return goerror.NewBusiness("mfa factor totp not found", goerror.CodeNotFound)
	// }

	// if factor.IsVerified {
	// 	slog.WarnContext(ctx, "mfa factor already verified", "user_id", cu.UserID, "mfa_id", mfaID)
	// 	return goerror.NewBusiness("mfa factor already verified", goerror.CodeConflict)
	// }

	// secretBytes, err := s.mfaCrypto.Decrypt(factor.Secret, mfacrypto.Scope{
	// 	UserID:  cu.UserID,
	// 	Purpose: mfacrypto.PurposeOTPSeed,
	// })
	// if err != nil {
	// 	slog.ErrorContext(ctx, "failed to decrypt totp secret", "user_id", cu.UserID, "mfa_id", mfaID, "error", err)
	// 	return goerror.NewServer(err)
	// }

	// if !s.totp.Validate(in.Code, string(secretBytes), s.clock.Now()) {
	// 	slog.WarnContext(ctx, "invalid totp code", "user_id", cu.UserID, "mfa_id", mfaID)
	// 	return goerror.NewBusiness("invalid code session", goerror.CodeUnauthorized)
	// }

	// if err := s.repoDB.MfaFactorVerify(ctx, factor.ID, cu.UserID); err != nil {
	// 	slog.ErrorContext(ctx, "failed to verify mfa factor", "user_id", cu.UserID, "mfa_id", mfaID, "error", err)
	// 	return goerror.NewServer(err)
	// }

	return nil
}
