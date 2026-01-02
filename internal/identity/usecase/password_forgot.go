package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type PasswordForgotInput struct {
	Email string `validate:"required,email"`
}

func (s *Usecase) PasswordForgot(ctx context.Context, in PasswordForgotInput) error {
	ctx, span := s.startSpan(ctx, "PasswordForgot")
	defer span.End()

	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	user, err := s.repoDB.GetUserByEmail(ctx, in.Email, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "password reset requested for unavailable user", "email", in.Email)
		return nil
	}
	if err != nil {
		slog.WarnContext(ctx, "failed to repo get user by email", "email", in.Email, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.ensureUserStatusAllowed(ctx, user.ID, user.Status); err != nil {
		slog.WarnContext(ctx, "password reset requested for ineligible user", "user_id", user.ID, "status", user.Status.String(), "error", err)
		return nil
	}

	cToken := s.oid.Generate()
	cTokenHash, err := s.hmac.Hash(cToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token", "error", err)
		return goerror.NewServer(err)
	}

	if err := s.repoDB.CreateChallenge(ctx, entity.Challenge{
		ID:        s.uid.Generate(),
		UserID:    user.ID,
		Token:     string(cTokenHash),
		Purpose:   entity.ChallengePurposePasswordForgotReset,
		ExpiresAt: s.clock.Now().Add(s.cfg.GetHour("modules.identity.password_reset_ttl_hours")),
	}); err != nil {
		slog.ErrorContext(ctx, "failed to repo create password reset challenge", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.repoMessaging.PublishUserForgotPassword(ctx, UserForgotPasswordEvent{
		UserID:         user.ID,
		Email:          user.Email,
		ChallengeToken: cToken,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to publish user forgot password", "user_id", user.ID, "error", err)
	}

	return nil
}
