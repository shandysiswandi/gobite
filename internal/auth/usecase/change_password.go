package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) ChangePassword(ctx context.Context, in domain.ChangePasswordInput) error {
	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	user, err := s.getUserByID(ctx, clm.Payload().UserID)
	if err != nil {
		return err
	}

	cred, err := s.getCredential(ctx, user.ID)
	if err != nil {
		return err
	}

	if !s.hash.Verify(cred.Password, in.CurrentPassword) {
		slog.WarnContext(ctx, "current password mismatch", "user_id", user.ID)
		return pkgerror.NewBusiness("invalid email or password", pkgerror.CodeUnauthorized)
	}

	newHash, err := s.hash.Hash(in.NewPassword)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash new password", "user_id", user.ID, "error", err)
		return pkgerror.NewServer(err)
	}

	if err := s.repoDB.UserCredentialUpdate(ctx, user.ID, string(newHash)); err != nil {
		slog.ErrorContext(ctx, "failed to update user password", "user_id", user.ID, "error", err)
		return pkgerror.NewServer(err)
	}

	return nil
}
