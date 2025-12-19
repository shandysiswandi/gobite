package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

func (s *Usecase) CountUnread(ctx context.Context, in CountUnreadInput) (*CountUnreadOutput, error) {
	if err := s.validator.Validate(in); err != nil {
		return nil, goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	uid := clm.GetInt64(keyPayloadUserID)

	count, err := s.repoDB.NotificationCountUnread(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo count unread notifications", "user_id", uid, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &CountUnreadOutput{Count: count}, nil
}
