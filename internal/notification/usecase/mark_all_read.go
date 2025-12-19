package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

func (s *Usecase) MarkAllRead(ctx context.Context, in MarkAllReadInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	uid := clm.GetInt64(keyPayloadUserID)

	if err := s.repoDB.NotificationsMarkAllRead(ctx, uid); err != nil {
		slog.ErrorContext(ctx, "failed to repo mark all notifications read", "user_id", uid, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
