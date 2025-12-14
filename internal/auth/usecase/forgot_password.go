package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

func (s *Usecase) ForgotPassword(ctx context.Context, in domain.ForgotPasswordInput) (*domain.ForgotPasswordOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, pkgerror.NewInvalidInput(err)
	}

	user, err := s.repoDB.UserGetByEmail(ctx, in.Email)
	if errors.Is(err, pkgerror.ErrNotFound) ||
		user.Status == domain.UserStatusBanned ||
		user.Status == domain.UserStatusUnverified {

		slog.WarnContext(ctx, "password reset requested for unavailable user", "email", in.Email, "error", err)
		return &domain.ForgotPasswordOutput{Success: true}, nil
	}
	if err != nil {
		slog.WarnContext(ctx, "failed to repo get user by email")
		return nil, pkgerror.NewServer(err)
	}

	ttl := time.Duration(s.cfg.GetInt("modules.auth.password_reset_ttl")) * time.Minute
	expiresAt := s.clock.Now().Add(ttl)
	token := s.uuid.Generate()

	if err := s.repoDB.UserPasswordResetCreate(ctx, user.ID, token, expiresAt); err != nil {
		slog.ErrorContext(ctx, "failed to repo create user password reset", "user_id", user.ID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	return &domain.ForgotPasswordOutput{Success: true}, nil
}
