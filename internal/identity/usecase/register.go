package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type RegisterInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,password"`
	FullName string `validate:"required,min=5,max=100,alphaspace"`
}

func (s *Usecase) Register(ctx context.Context, in RegisterInput) error {
	ctx, span := s.startSpan(ctx, "Register")
	defer span.End()

	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.FullName = strings.TrimSpace(in.FullName)

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	user, err := s.repoDB.GetUserByEmail(ctx, in.Email, true)
	if err == nil {
		switch user.Status {
		case entity.UserStatusActive:
			return goerror.NewBusiness("Email already registered", goerror.CodeConflict)
		case entity.UserStatusUnverified:
			return goerror.NewBusiness("Account not verified", goerror.CodeConflict)
		case entity.UserStatusInactive:
			return goerror.NewBusiness("Account deactivated", goerror.CodeConflict)
		default:
			return goerror.NewBusiness("Account not allowed", goerror.CodeForbidden)
		}
	}
	if !errors.Is(err, goerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", in.Email, "error", err)
		return goerror.NewServer(err)
	}

	hashedPassword, err := s.bcrypt.Hash(in.Password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error", err)
		return goerror.NewServer(err)
	}

	newUserID := s.uid.Generate()
	newUser := entity.NewUser{
		ID:        newUserID,
		CreatedBy: newUserID,
		UpdatedBy: newUserID,
		Email:     in.Email,
		FullName:  in.FullName,
		AvatarURL: "https://ui-avatars.com/api/?name=" + url.QueryEscape(in.FullName),
		Status:    entity.UserStatusUnverified,
	}

	cToken := s.oid.Generate()
	cTokenHash, err := s.hmac.Hash(cToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
		return goerror.NewServer(err)
	}

	challenge := entity.Challenge{
		ID:        s.uid.Generate(),
		UserID:    newUser.ID,
		Token:     string(cTokenHash),
		Purpose:   entity.ChallengePurposeRegisterVerify,
		ExpiresAt: s.clock.Now().Add(s.cfg.GetHour("modules.identity.registration_ttl_hours")),
	}

	if err := s.repoDB.NewRegistration(ctx, newUser, challenge, string(hashedPassword)); err != nil {
		slog.ErrorContext(ctx, "failed to repo user registration", "email", newUser.Email, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.repoMessaging.PublishUserRegistration(ctx, UserRegistrationEvent{
		UserID:         newUser.ID,
		Email:          newUser.Email,
		FullName:       newUser.FullName,
		ChallengeToken: cToken,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to publish user registration", "user_id", newUser.ID, "error", err)
	}

	return nil
}
