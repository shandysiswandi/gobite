package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type RegisterResendInput struct {
	Email string `validate:"required,email"`
}

func (s *Usecase) RegisterResend(ctx context.Context, in RegisterResendInput) error {
	ctx, span := s.startSpan(ctx, "RegisterResend")
	defer span.End()

	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	user, err := s.repoDB.GetUserByEmail(ctx, in.Email, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "email not registered for resend", "email", in.Email)
		return nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", in.Email, "error", err)
		return goerror.NewServer(err)
	}

	if user.Status != entity.UserStatusUnverified {
		slog.WarnContext(ctx, "failed to process resend email", "user_id", user.ID, "status", user.Status.String())
		return nil
	}

	cToken := s.oid.Generate()
	cTokenHash, err := s.hmac.Hash(cToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return goerror.NewServer(err)
	}

	challenge := entity.Challenge{
		ID:        s.uid.Generate(),
		UserID:    user.ID,
		Token:     string(cTokenHash),
		Purpose:   entity.ChallengePurposeRegisterVerify,
		ExpiresAt: s.clock.Now().Add(s.cfg.GetHour("modules.identity.registration_ttl_hours")),
	}

	if err := s.repoDB.CreateChallenge(ctx, challenge); err != nil {
		slog.ErrorContext(ctx, "failed to repo create register challenge", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.repoMessaging.PublishUserRegistration(ctx, UserRegistrationEvent{
		UserID:         user.ID,
		Email:          user.Email,
		FullName:       user.FullName,
		ChallengeToken: cToken,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to publish user registration resend", "user_id", user.ID, "error", err)
	}

	return nil
}
