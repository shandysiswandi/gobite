package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

func (s *Usecase) MarkRead(ctx context.Context, in MarkReadInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	uid := clm.GetInt64(keyPayloadUserID)

	if err := s.repoDB.NotificationMarkRead(ctx, uid, in.NotificationID); err != nil {
		slog.ErrorContext(ctx, "failed to repo mark notification read", "user_id", uid, "notification_id", in.NotificationID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
