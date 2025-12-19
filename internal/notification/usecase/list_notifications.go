package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

func (s *Usecase) ListNotifications(ctx context.Context, in ListNotificationsInput) (*ListNotificationsOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	uid := clm.GetInt64(keyPayloadUserID)

	items, err := s.repoDB.NotificationGetUserNotificationPaginate(ctx, uid, in.Limit, in.Offset)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user notifications paginate", "user_id", uid, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &ListNotificationsOutput{
		Notifications: items,
		Limit:         in.Limit,
		Offset:        in.Offset,
	}, nil
}
