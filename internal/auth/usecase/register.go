package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
)

func (s *Usecase) Register(ctx context.Context, in domain.RegisterInput) (*domain.RegisterOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	fullName := strings.TrimSpace(in.FullName)

	_, err := s.repoDB.UserGetByEmail(ctx, email)
	if err == nil {
		slog.WarnContext(ctx, "email already registered")
		return nil, pkgerror.ErrAuthEmailUsed
	}
	if !errors.Is(err, pkgerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user by email", "error", err)
		return nil, err
	}

	hashedPassword, err := s.hash.Hash(in.Password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password")
		return nil, err
	}

	user := domain.User{
		ID:        s.uid.Generate(),
		Email:     email,
		FullName:  fullName,
		AvatarURL: "",
		Status:    domain.UserStatusUnverified,
	}

	if err := s.repoDB.UserRegistration(ctx, user, string(hashedPassword)); err != nil {
		slog.ErrorContext(ctx, "failed to repo user registration", "user_id", user.ID, "error", err)
		return nil, err
	}

	return &domain.RegisterOutput{IsNeedVerify: true}, nil
}
