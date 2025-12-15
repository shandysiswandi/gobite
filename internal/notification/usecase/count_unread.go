package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) CountUnread(ctx context.Context, in CountUnreadInput) (*CountUnreadOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, pkgerror.NewInvalidInput(err)
	}

	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	count, err := s.repoDB.NotificationCountUnread(ctx, clm.Payload().UserID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo count unread notifications", "user_id", clm.Payload().UserID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	return &CountUnreadOutput{Count: count}, nil
}
