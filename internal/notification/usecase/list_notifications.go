package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type ListInboxInput struct {
	Status string `validate:"omitempty,oneof=all unread read"`
	Limit  int32  `validate:"omitempty,gte=1,lte=100"`
	Offset int32  `validate:"omitempty,gte=0"`
}

func (s *Usecase) ListInbox(ctx context.Context, in ListInboxInput) (_ []entity.NotificationItem, err error) {
	ctx, span := s.startSpan(ctx, "ListInbox")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	if in.Status == "" {
		in.Status = string(entity.NotificationStatusAll)
	}
	if in.Limit == 0 {
		in.Limit = 20
	}

	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	items, err := s.repoDB.ListNotifications(ctx, clm.UserID, entity.NotificationStatus(in.Status), in.Limit, in.Offset)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo list notifications", "user_id", clm.UserID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return items, nil
}
