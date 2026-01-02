package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type ProfileUpdateInput struct {
	FullName string `validate:"required,min=5,max=100,alphaspace"`
}

func (s *Usecase) ProfileUpdate(ctx context.Context, in ProfileUpdateInput) error {
	ctx, span := s.startSpan(ctx, "ProfileUpdate")
	defer span.End()

	in.FullName = strings.TrimSpace(in.FullName)

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	user, err := s.repoDB.GetUserByEmail(ctx, clm.UserEmail, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "email", clm.UserEmail)
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", clm.UserEmail, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.ensureUserStatusAllowed(ctx, user.ID, user.Status); err != nil {
		return err
	}

	if err := s.repoDB.UpdateUserProfile(ctx, user.ID, in.FullName); err != nil {
		slog.ErrorContext(ctx, "failed to update user profile", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
