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

	email := strings.TrimSpace(strings.ToLower(in.Email))
	fullName := strings.TrimSpace(in.FullName)

	in = RegisterInput{
		Email:    email,
		Password: in.Password,
		FullName: fullName,
	}

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	user, err := s.repoDB.GetUserByEmail(ctx, email, true)
	if err == nil {
		switch user.Status {
		case entity.UserStatusActive, entity.UserStatusUnverified, entity.UserStatusDeleted:
			slog.WarnContext(ctx, "email already registered", "email", email)
			return goerror.NewBusiness("email is already registered", goerror.CodeConflict)
		default:
			slog.WarnContext(ctx, "user is not eligible", "email", email)
			return goerror.NewBusiness("user cannot be registered", goerror.CodeForbidden)
		}
	}
	if !errors.Is(err, goerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", email, "error", err)
		return goerror.NewServer(err)
	}

	hashedPassword, err := s.bcrypt.Hash(in.Password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error", err)
		return goerror.NewServer(err)
	}

	newUser := entity.User{
		ID:        s.uid.Generate(),
		Email:     email,
		FullName:  fullName,
		AvatarURL: "https://ui-avatars.com/api/?name=" + url.QueryEscape(fullName),
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
