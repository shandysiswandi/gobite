package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type BackupCodeRotateInput struct {
	CurrentPassword string `validate:"required"`
}

type BackupCodeRotateOutput struct {
	RecoveryCodes []string
}

func (s *Usecase) BackupCodeRotate(ctx context.Context, in BackupCodeRotateInput) (*BackupCodeRotateOutput, error) {
	ctx, span := s.startSpan(ctx, "BackupCodeRotate")
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

	if err := s.repoDB.NewBackupCodes(ctx, codes); err != nil {
		slog.ErrorContext(ctx, "failed to rotate backup codes", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &BackupCodeRotateOutput{RecoveryCodes: recoveryCodes}, nil
}
