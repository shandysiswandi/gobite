package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type MarkNotificationReadInput struct {
	ID int64 `validate:"required,gt=0"`
}

func (s *Usecase) MarkNotificationRead(ctx context.Context, in MarkNotificationReadInput) error {
	ctx, span := s.startSpan(ctx, "MarkNotificationRead")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return err
	}

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	updated, err := s.repoDB.MarkNotificationRead(ctx, clm.UserID, in.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo mark notification read", "user_id", clm.UserID, "notification_id", in.ID, "error", err)
		return goerror.NewServer(err)
	}
	if !updated {
		return goerror.NewBusiness("notification not found", goerror.CodeNotFound)
	}

	return nil
}

func (s *Usecase) MarkAllNotificationsRead(ctx context.Context) error {
	ctx, span := s.startSpan(ctx, "MarkAllNotificationsRead")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return err
	}

	if _, err := s.repoDB.MarkNotificationsReadAll(ctx, clm.UserID); err != nil {
		slog.ErrorContext(ctx, "failed to repo mark all notifications read", "user_id", clm.UserID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}

type DeleteNotificationInput struct {
	ID int64 `validate:"required,gt=0"`
}

func (s *Usecase) DeleteNotification(ctx context.Context, in DeleteNotificationInput) error {
	ctx, span := s.startSpan(ctx, "DeleteNotification")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return err
	}

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	deleted, err := s.repoDB.SoftDeleteNotification(ctx, clm.UserID, in.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo delete notification", "user_id", clm.UserID, "notification_id", in.ID, "error", err)
		return goerror.NewServer(err)
	}
	if !deleted {
		return goerror.NewBusiness("notification not found", goerror.CodeNotFound)
	}

	return nil
}
