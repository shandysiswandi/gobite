package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mfa"
)

type TOTPConfirmInput struct {
	ChallengeToken string `validate:"required"`
	Code           string `validate:"required,len=6,numeric"`
}

type TOTPConfirmOutput struct {
	RecoveryCodes []string
}

func (s *Usecase) TOTPConfirm(ctx context.Context, in TOTPConfirmInput) (*TOTPConfirmOutput, error) {
	ctx, span := s.startSpan(ctx, "TOTPConfirm")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	cTokenHash, err := s.hmac.Hash(in.ChallengeToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return nil, goerror.NewServer(err)
	}

	cu, err := s.getChallengeUserByToken(ctx, string(cTokenHash))
	if err != nil {
		return nil, err
	}

	friendlyName, keyVersion, err := s.validateChallengeUser(ctx, cu, clm.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.ensureNoTOTPFactor(ctx, cu.UserID); err != nil {
		return nil, err
	}

	secretCiphertext, err := s.decodeTOTPSecret(ctx, cu)
	if err != nil {
		return nil, err
	}

	secretBytes, err := s.decryptTOTPSecret(ctx, cu, secretCiphertext)
	if err != nil {
		return nil, err
	}

	if !s.totp.Validate(in.Code, string(secretBytes), s.clock.Now()) {
		slog.WarnContext(ctx, "invalid totp code", "user_id", cu.UserID, "challenge_id", cu.ChallengeID)
		return nil, goerror.NewBusiness("invalid code session", goerror.CodeUnauthorized)
	}

	factorTotp, factorBackupCode := s.buildTOTPFacts(cu, friendlyName, keyVersion, secretCiphertext)

	recoveryCodes, err := s.mfaRecoveryCode.Generate()
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate recovery codes", "user_id", cu.UserID, "challenge_id", cu.ChallengeID, "error", err)
		return nil, goerror.NewServer(err)
	}

	codes, err := s.buildBackupCodes(ctx, cu, recoveryCodes)
	if err != nil {
		return nil, err
	}

	if err := s.repoDB.NewMFAFactorTOTPBackup(ctx, factorTotp, factorBackupCode, codes, cu.ChallengeID); err != nil {
		slog.ErrorContext(ctx, "failed to repo new mfa factor totp and backup code", "user_id", cu.UserID, "challenge_id", cu.ChallengeID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &TOTPConfirmOutput{RecoveryCodes: recoveryCodes}, nil
}

func (s *Usecase) getChallengeUserByToken(ctx context.Context, tokenHash string) (*entity.ChallengeUser, error) {
	cu, err := s.repoDB.GetChallengeUserByTokenPurpose(ctx, tokenHash, entity.ChallengePurposeMFASetupConfirm)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "challenge user not found", "challenge_token", tokenHash)
		return nil, goerror.NewBusiness("invalid challenge session", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get challange user by token purpose", "challenge_token", tokenHash, "error", err)
		return nil, goerror.NewServer(err)
	}
	return cu, nil
}

func (s *Usecase) validateChallengeUser(ctx context.Context, cu *entity.ChallengeUser, userID int64) (string, int, error) {
	if err := s.ensureUserStatusAllowed(ctx, cu.UserID, cu.UserStatus); err != nil {
		return "", 0, err
	}

	if cu.UserID != userID {
		slog.WarnContext(ctx, "challenge user mismatch", "user_id", userID, "challenge_user_id", cu.UserID)
		return "", 0, goerror.NewBusiness("invalid challenge session", goerror.CodeUnauthorized)
	}

	secretEncoded := cu.ChallengeMetadata.GetString("secret")
	if secretEncoded == "" {
		slog.WarnContext(ctx, "challenge missing totp secret", "user_id", cu.UserID, "challenge_id", cu.ChallengeID)
		return "", 0, goerror.NewBusiness("invalid challenge session", goerror.CodeUnauthorized)
	}

	friendlyName := strings.TrimSpace(cu.ChallengeMetadata.GetString("friendly_name"))
	if friendlyName == "" {
		slog.WarnContext(ctx, "challenge missing totp friendly name", "user_id", cu.UserID, "challenge_id", cu.ChallengeID)
		return "", 0, goerror.NewBusiness("invalid challenge session", goerror.CodeUnauthorized)
	}

	keyVersion := cu.ChallengeMetadata.GetInt("key_version")
	if keyVersion == 0 {
		keyVersion = 1
	}

	return friendlyName, keyVersion, nil
}

func (s *Usecase) ensureNoTOTPFactor(ctx context.Context, userID int64) error {
	verifiedFactors, err := s.repoDB.GetMFAFactorByUserID(ctx, userID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get verified mfa factor", "user_id", userID, "error", err)
		return goerror.NewServer(err)
	}

	for i := range verifiedFactors {
		if verifiedFactors[i].Type == entity.MFATypeTOTP {
			return goerror.NewBusiness("A verified TOTP factor already exists", goerror.CodeConflict)
		}
	}

	return nil
}

func (s *Usecase) decodeTOTPSecret(ctx context.Context, cu *entity.ChallengeUser) ([]byte, error) {
	secretEncoded := cu.ChallengeMetadata.GetString("secret")
	secretCiphertext, err := base64.StdEncoding.DecodeString(secretEncoded)
	if err != nil {
		slog.WarnContext(ctx, "challenge totp secret decode failed", "user_id", cu.UserID, "challenge_id", cu.ChallengeID, "error", err)
		return nil, goerror.NewBusiness("invalid challenge session", goerror.CodeUnauthorized)
	}
	return secretCiphertext, nil
}

func (s *Usecase) decryptTOTPSecret(ctx context.Context, cu *entity.ChallengeUser, secretCiphertext []byte) ([]byte, error) {
	secretBytes, err := s.mfaEncryptor.Decrypt(secretCiphertext, mfa.Scope{
		UserID:  cu.UserID,
		Purpose: mfa.PurposeOTPSeed,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to decrypt totp secret", "user_id", cu.UserID, "challenge_id", cu.ChallengeID, "error", err)
		return nil, goerror.NewServer(err)
	}
	return secretBytes, nil
}

func (s *Usecase) buildTOTPFacts(cu *entity.ChallengeUser, friendlyName string, keyVersion int, secretCiphertext []byte) (entity.MFAFactor, entity.MFAFactor) {
	factorTotp := entity.MFAFactor{
		ID:           s.uid.Generate(),
		UserID:       cu.UserID,
		Type:         entity.MFATypeTOTP,
		FriendlyName: friendlyName,
		Secret:       secretCiphertext,
		KeyVersion:   int16(keyVersion),
		IsVerified:   true,
	}
	factorBackupCode := entity.MFAFactor{
		ID:           s.uid.Generate(),
		UserID:       cu.UserID,
		Type:         entity.MFATypeBackupCode,
		FriendlyName: friendlyName,
		Secret:       []byte(""),
		KeyVersion:   int16(keyVersion),
		IsVerified:   true,
	}
	return factorTotp, factorBackupCode
}

func (s *Usecase) buildBackupCodes(ctx context.Context, cu *entity.ChallengeUser, recoveryCodes []string) ([]entity.MFABackupCode, error) {
	codes := make([]entity.MFABackupCode, 0, len(recoveryCodes))
	for _, code := range recoveryCodes {
		hashed, err := s.argon2id.Hash(code)
		if err != nil {
			slog.ErrorContext(ctx, "failed to hash recovery code", "user_id", cu.UserID, "challenge_id", cu.ChallengeID, "error", err)
			return nil, goerror.NewServer(err)
		}

		codes = append(codes, entity.MFABackupCode{
			ID:     s.uid.Generate(),
			UserID: cu.UserID,
			Code:   string(hashed),
		})
	}
	return codes, nil
}
