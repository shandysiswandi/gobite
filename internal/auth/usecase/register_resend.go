package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type RegisterResendInput struct {
	Email string `validate:"required,lowercase,email"`
	IP    string
}

func (s *Usecase) RegisterResend(ctx context.Context, in RegisterResendInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	email := strings.TrimSpace(in.Email)
	allowed, err := s.repoCache.RegisterResendAllow(ctx, "user:register:"+email, 2*time.Minute)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo cache register resend allow", "email", email, "error", err)
		return goerror.NewServer(err)
	}

	if !allowed {
		slog.WarnContext(ctx, "resend blocked, rate limit reach", "email", email)
		return nil
	}

	user, err := s.repoDB.GetUserByEmail(ctx, email, false)
	if errors.Is(err, goerror.ErrNotFound) {
		// PROTECTION (User Enumeration): success even if user doesn't exist.
		slog.WarnContext(ctx, "email not registered for resend", "email", email)
		return nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", email, "error", err)
		return goerror.NewServer(err)
	}

	uStatus := user.Status.Ensure()
	if uStatus == entity.UserStatusUnverified {
		if err := s.repoMessaging.PublishUserRegistration(ctx, entity.UserRegistrationMessage{
			UserID:   user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to publish user registration resend", "user_id", user.ID, "error", err)
		}

		return nil
	}

	// PROTECTION (User Enumeration): success even if user status not valid.
	slog.WarnContext(ctx, "failed to process resend email", "email", email, "status", uStatus.String())
	return nil
}
