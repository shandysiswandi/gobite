package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mfacrypto"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type SetupTOTPInput struct {
	FriendlyName    string `validate:"required,min=2,max=100"`
	CurrentPassword string `validate:"required"`
}

type SetupTOTPOutput struct {
	ChallengeToken string
	Key            string
	URI            string
}

func (s *Usecase) SetupTOTP(ctx context.Context, in SetupTOTPInput) (*SetupTOTPOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	userID := clm.GetInt64(keyPayloadUserID)
	user, err := s.repoDB.GetUserCredentialInfo(ctx, clm.GetInt64(keyPayloadUserID))
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", userID)
		return nil, goerror.NewBusiness("user account not found", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if !s.password.Verify(user.Password, in.CurrentPassword) {
		slog.WarnContext(ctx, "password user account not match", "user_id", userID)
		return nil, goerror.NewBusiness("invalid password", goerror.CodeUnauthorized)
	}

	if err := s.ensureUserAllowedToLogin(ctx, user.ID, user.Status); err != nil {
		return nil, err
	}

	verifiedFactors, err := s.repoDB.GetMFAFactorByUserID(ctx, user.ID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get verified mfa factor", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	isMFATOTPVerifiedExist := false
	for _, f := range verifiedFactors {
		if f.Type == entity.MFATypeTOTP {
			isMFATOTPVerifiedExist = true
			break
		}
	}

	if isMFATOTPVerifiedExist {
		return nil, goerror.NewBusiness("A verified TOTP factor already exists", goerror.CodeConflict)
	}

	secret, uri, err := s.totp.Generate(user.Email)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate totp secret", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	encryptedSecret, err := s.mfaCrypto.Encrypt([]byte(secret), mfacrypto.Scope{
		UserID:  user.ID,
		Purpose: mfacrypto.PurposeOTPSeed,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to encrypt totp secret", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	cToken := s.oid.Generate()
	cTokenHash, err := s.hash.Hash(cToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return nil, goerror.NewServer(err)
	}

	factor := entity.MFAFactor{
		ID:           s.uid.Generate(),
		UserID:       user.ID,
		Type:         entity.MFATypeTOTP,
		FriendlyName: strings.TrimSpace(in.FriendlyName),
		Secret:       encryptedSecret,
		KeyVersion:   1, // can be use config later
		IsVerified:   false,
	}
	challenge := entity.Challenge{
		ID:        s.uid.Generate(),
		UserID:    user.ID,
		Token:     string(cTokenHash),
		Purpose:   entity.ChallengePurposeMFASetupConfirm,
		ExpiresAt: s.clock.Now().Add(5 * time.Minute), // can be use config later
		Metadata:  valueobject.JSONMap{"mfa_id": factor.ID},
	}

	if err := s.repoDB.CreateMFAFactorAndChallenge(ctx, factor, challenge); err != nil {
		slog.ErrorContext(ctx, "failed to create mfa factor and challenge", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &SetupTOTPOutput{
		ChallengeToken: cToken,
		Key:            secret,
		URI:            uri,
	}, nil
}
