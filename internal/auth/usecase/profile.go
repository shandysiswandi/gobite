package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type ProfileInput struct{}

type ProfileOutput struct {
	ID        int64
	Email     string
	FullName  string
	AvatarURL string
	Status    string
}

func (s *Usecase) Profile(ctx context.Context, in ProfileInput) (*ProfileOutput, error) {
	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	userID := clm.GetString(keyPayloadUserEmail)
	user, err := s.repoDB.GetUserByEmail(ctx, userID, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", userID)
		return nil, goerror.NewBusiness("invalid email or password", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
		return nil, goerror.NewServer(err)
	}

	if err := s.ensureUserAllowedToLogin(ctx, user.ID, user.Status); err != nil {
		return nil, err
	}

	return &ProfileOutput{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status.String(),
	}, nil
}
