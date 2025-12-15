package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) MarkAllRead(ctx context.Context, in MarkAllReadInput) error {
	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	if err := s.repoDB.NotificationsMarkAllRead(ctx, clm.Payload().UserID); err != nil {
		slog.ErrorContext(ctx, "failed to repo mark all notifications read", "user_id", clm.Payload().UserID, "error", err)
		return pkgerror.NewServer(err)
	}

	return nil
}
