package usecase

import (
	"context"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

type (
	DeviceRemoveInput struct {
		DeviceToken string `validate:"required"`
	}
)

func (s *Usecase) DeviceRemove(ctx context.Context, in DeviceRemoveInput) error {
	ctx, span := s.startSpan(ctx, "DeviceRemove")
	defer span.End()

	in.DeviceToken = strings.TrimSpace(in.DeviceToken)

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	if err := s.repoDB.RemoveUserDevice(ctx, in.DeviceToken); err != nil {
		slog.ErrorContext(ctx, "failed to repo remove device token", "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
