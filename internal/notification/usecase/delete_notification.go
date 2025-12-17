package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) DeleteNotification(ctx context.Context, in DeleteNotificationInput) error {
	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	if err := s.repoDB.NotificationSoftDelete(ctx, clm.Payload().UserID, in.NotificationID); err != nil {
		slog.ErrorContext(ctx, "failed to repo soft delete notification", "user_id", clm.Payload().UserID, "notification_id", in.NotificationID, "error", err)
		return pkgerror.NewServer(err)
	}

	return nil
}
