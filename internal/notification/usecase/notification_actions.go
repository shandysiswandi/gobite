package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type MarkInboxReadInput struct {
	ID int64 `validate:"required,gt=0"`
}

func (s *Usecase) MarkInboxRead(ctx context.Context, in MarkInboxReadInput) error {
	ctx, span := s.startSpan(ctx, "MarkInboxRead")
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
		slog.ErrorContext(ctx, "failed to repo mark inbox read", "user_id", clm.UserID, "notification_id", in.ID, "error", err)
		return goerror.NewServer(err)
	}
	if !updated {
		return goerror.NewBusiness("inbox notification not found", goerror.CodeNotFound)
	}

	return nil
}

func (s *Usecase) MarkAllInboxRead(ctx context.Context) error {
	ctx, span := s.startSpan(ctx, "MarkAllInboxRead")
	defer span.End()

	clm, err := s.requireAuth(ctx)
	if err != nil {
		return err
	}

	if _, err := s.repoDB.MarkNotificationsReadAll(ctx, clm.UserID); err != nil {
		slog.ErrorContext(ctx, "failed to repo mark all inbox read", "user_id", clm.UserID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}

type DeleteInboxInput struct {
	ID int64 `validate:"required,gt=0"`
}

func (s *Usecase) DeleteInbox(ctx context.Context, in DeleteInboxInput) error {
	ctx, span := s.startSpan(ctx, "DeleteInbox")
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
		slog.ErrorContext(ctx, "failed to repo delete inbox notification", "user_id", clm.UserID, "notification_id", in.ID, "error", err)
		return goerror.NewServer(err)
	}
	if !deleted {
		return goerror.NewBusiness("inbox notification not found", goerror.CodeNotFound)
	}

	return nil
}
