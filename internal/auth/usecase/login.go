package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/shandysiswandi/gobite/internal/auth/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type LoginOutput struct {
	MfaRequired    bool
	ChallengeToken string
	//
	AccessToken  string
	RefreshToken string
}

func (s *Usecase) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	email := strings.TrimSpace(in.Email)
	user, err := s.repoDB.GetUserLoginInfo(ctx, email)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "email", email)
		return nil, goerror.NewBusiness("invalid email or password", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", email, "error", err)
		return nil, goerror.NewServer(err)
	}

	if err := s.ensureUserAllowedToLogin(ctx, user.ID, user.Status); err != nil {
		return nil, err
	}

	if !s.password.Verify(user.Password, in.Password) {
		slog.WarnContext(ctx, "password user account not match", "user_id", user.ID)
		return nil, goerror.NewBusiness("invalid email or password", goerror.CodeUnauthorized)
	}

	if user.HasMFA {
		cToken := s.oid.Generate()

		cTokenHash, err := s.hash.Hash(cToken)
		if err != nil {
			slog.ErrorContext(ctx, "failed to hash token challange", "error", err)
			return nil, goerror.NewServer(err)
		}

		if err := s.repoDB.CreateChallenge(ctx, entity.Challenge{
			ID:        s.uid.Generate(),
			UserID:    user.ID,
			Token:     string(cTokenHash),
			Purpose:   entity.ChallengePurposeMFALogin,
			ExpiresAt: s.clock.Now().Add(5 * time.Minute), // can be use config later
		}); err != nil {
			slog.ErrorContext(ctx, "failed to repo create challange", "user_id", user.ID, "error", err)
			return nil, goerror.NewServer(err)
		}

		return &LoginOutput{
			MfaRequired:    true,
			ChallengeToken: cToken,
		}, nil
	}

	acToken, err := s.jwt.Generate(
		jwt.WithID(s.uuid.Generate()),
		jwt.WithSubject(strconv.FormatInt(user.ID, 10)),
		jwt.WithPayloadValue(keyPayloadUserID, user.ID),
		jwt.WithPayloadValue(keyPayloadUserEmail, user.Email),
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate access jwt token", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	refToken := s.oid.Generate()
	refTokenHash, err := s.hash.Hash(refToken)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash refresh token", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if err := s.repoDB.CreateRefreshToken(ctx, entity.RefreshToken{
		ID:        s.uid.Generate(),
		UserID:    user.ID,
		Token:     string(refTokenHash),
		ExpiresAt: s.clock.Now().Add(time.Duration(s.cfg.GetInt("modules.auth.refresh_token_ttl")) * 24 * time.Hour),
	}); err != nil {
		slog.ErrorContext(ctx, "failed to repo create refresh token user", "user_id", user.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &LoginOutput{
		AccessToken:  acToken,
		RefreshToken: refToken,
	}, nil
}
