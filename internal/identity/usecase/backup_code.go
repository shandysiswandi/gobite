package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type BackupCodeInput struct {
	CurrentPassword string `validate:"required"`
}

type BackupCodeOutput struct {
	RecoveryCodes []string
}

func (s *Usecase) BackupCode(ctx context.Context, in BackupCodeInput) (*BackupCodeOutput, error) {
	ctx, span := s.startSpan(ctx, "BackupCode")
	defer span.End()

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

	factor, err := s.ensureBackupCodeFactor(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	recoveryCodes, err := s.mfaRecoveryCode.Generate()
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate backup codes", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	codes := make([]entity.MFABackupCode, 0, len(recoveryCodes))
	for _, code := range recoveryCodes {
		hashed, err := s.argon2id.Hash(code)
		if err != nil {
			slog.ErrorContext(ctx, "failed to hash backup code", "user_id", user.ID, "error", err)
			return nil, goerror.NewServer(err)
		}

		codes = append(codes, entity.MFABackupCode{
			ID:     s.uid.Generate(),
			UserID: user.ID,
			Code:   string(hashed),
		})
	}

	if err := s.repoDB.NewBackupCodes(ctx, user.ID, codes, factor); err != nil {
		slog.ErrorContext(ctx, "failed to rotate backup codes", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &BackupCodeOutput{RecoveryCodes: recoveryCodes}, nil
}

func (s *Usecase) ensureBackupCodeFactor(ctx context.Context, userID int64) (*entity.MFAFactor, error) {
	hasFactor, err := s.hasBackupCodeFactor(ctx, userID)
	if err != nil {
		return nil, err
	}
	if hasFactor {
		return nil, nil
	}

	factor := &entity.MFAFactor{
		ID:           s.uid.Generate(),
		UserID:       userID,
		Type:         entity.MFATypeBackupCode,
		FriendlyName: "Backup Codes",
		Secret:       []byte(""),
		KeyVersion:   1,
		IsVerified:   true,
	}

	return factor, nil
}

func (s *Usecase) hasBackupCodeFactor(ctx context.Context, userID int64) (bool, error) {
	factors, err := s.repoDB.GetMFAFactorByUserID(ctx, userID, true)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get verified mfa factor", "user_id", userID, "error", err)
		return false, goerror.NewServer(err)
	}
	for i := range factors {
		if factors[i].Type == entity.MFATypeBackupCode {
			return true, nil
		}
	}

	return false, nil
}
