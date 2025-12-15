package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) MarkRead(ctx context.Context, in MarkReadInput) error {
	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	if err := s.repoDB.NotificationMarkRead(ctx, clm.Payload().UserID, in.NotificationID); err != nil {
		slog.ErrorContext(ctx, "failed to repo mark notification read", "user_id", clm.Payload().UserID, "notification_id", in.NotificationID, "error", err)
		return pkgerror.NewServer(err)
	}

	return nil
}
