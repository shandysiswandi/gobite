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
		return nil, pkgerror.NewInvalidInput(err)
	}

	_, err := s.repoDB.UserGetByEmail(ctx, in.Email)
	if err == nil {
		slog.WarnContext(ctx, "email already registered")
		return nil, pkgerror.NewBusiness("email is already registered", pkgerror.CodeConflict)
	}
	if !errors.Is(err, pkgerror.ErrNotFound) {
		slog.ErrorContext(ctx, "failed to repo get user by email", "error", err)
		return nil, pkgerror.NewServer(err)
	}

	hashedPassword, err := s.hash.Hash(in.Password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error", err)
		return nil, pkgerror.NewServer(err)
	}

	user := domain.User{
		ID:        s.uid.Generate(),
		Email:     in.Email,
		FullName:  strings.TrimSpace(in.FullName),
		AvatarURL: "",
		Status:    domain.UserStatusUnverified,
	}

	if err := s.repoDB.UserRegistration(ctx, user, string(hashedPassword)); err != nil {
		slog.ErrorContext(ctx, "failed to repo user registration", "user_id", user.ID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	return &domain.RegisterOutput{IsNeedVerify: true}, nil
}
