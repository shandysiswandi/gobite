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
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type TOTPSetupInput struct {
	FriendlyName    string `validate:"required,min=2,max=100"`
	CurrentPassword string `validate:"required"`
}

type TOTPSetupOutput struct {
	ChallengeToken string
	Key            string
	URI            string
}

func (s *Usecase) TOTPSetup(ctx context.Context, in TOTPSetupInput) (*TOTPSetupOutput, error) {
	ctx, span := s.startSpan(ctx, "TOTPSetup")
	defer span.End()

	in.FriendlyName = strings.TrimSpace(in.FriendlyName)
	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	user, err := s.repoDB.GetUserCredentialInfo(ctx, clm.UserID)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", clm.UserID)
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", clm.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if !s.bcrypt.Verify(user.Password, in.CurrentPassword) {
		slog.WarnContext(ctx, "password user account not match", "user_id", user.ID)
		return nil, goerror.NewBusiness("invalid password", goerror.CodeUnauthorized)
	}

	if err := s.ensureUserStatusAllowed(ctx, user.ID, user.Status); err != nil {
		return nil, err
	}

	verifiedFactors, err := s.repoDB.GetMFAFactorByUserID(ctx, user.ID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get verified mfa factor", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	isMFATOTPVerifiedExist := false
	for i := range verifiedFactors {
		if verifiedFactors[i].Type == entity.MFATypeTOTP {
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

	encryptedSecret, err := s.mfaEncryptor.Encrypt([]byte(secret), mfa.Scope{
		UserID:  user.ID,
		Purpose: mfa.PurposeOTPSeed,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to encrypt totp secret", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	cToken := s.oid.Generate()
	cTokenHash, err := s.hmac.Hash(cToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return nil, goerror.NewServer(err)
	}

	challenge := entity.Challenge{
		ID:        s.uid.Generate(),
		UserID:    user.ID,
		Token:     string(cTokenHash),
		Purpose:   entity.ChallengePurposeMFASetupConfirm,
		ExpiresAt: s.clock.Now().Add(s.cfg.GetMinute("modules.identity.mfa_setup_confirm_ttl_minutes")),
		Metadata: valueobject.JSONMap{
			"secret":        base64.StdEncoding.EncodeToString(encryptedSecret),
			"friendly_name": in.FriendlyName,
			"key_version":   1, // can be use config later
		},
	}

	if err := s.repoDB.CreateChallenge(ctx, challenge); err != nil {
		slog.ErrorContext(ctx, "failed to create mfa challenge", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &TOTPSetupOutput{
		ChallengeToken: cToken,
		Key:            secret,
		URI:            uri,
	}, nil
}
