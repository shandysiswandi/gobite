package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) Profile(ctx context.Context, in domain.ProfileInput) (*domain.ProfileOutput, error) {
	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	userID, err := strconv.ParseInt(clm.Subject(), 10, 64)
	if err != nil {
		slog.WarnContext(ctx, "invalid access token subject", "subject", clm.Subject(), "error", err)
		return nil, pkgerror.ErrUnauthenticated
	}

	user, err := s.repoDB.UserGetByID(ctx, userID)
	if errors.Is(err, pkgerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "user_id", userID)
		return nil, pkgerror.ErrUnauthenticated
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", userID, "error", err)
		return nil, err
	}

	return &domain.ProfileOutput{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
		Status:    user.Status.String(),
	}, nil
}
