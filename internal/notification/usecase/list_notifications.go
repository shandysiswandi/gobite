package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) ListNotifications(ctx context.Context, in ListNotificationsInput) (*ListNotificationsOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, pkgerror.NewInvalidInput(err)
	}

	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	items, err := s.repoDB.NotificationGetUserNotificationPaginate(ctx, clm.Payload().UserID, in.Limit, in.Offset)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user notifications paginate", "user_id", clm.Payload().UserID, "error", err)
		return nil, pkgerror.NewServer(err)
	}

	return &ListNotificationsOutput{
		Notifications: items,
		Limit:         in.Limit,
		Offset:        in.Offset,
	}, nil
}
