package usecase

import (
	"context"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgerror"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
)

func (s *Usecase) RegisterDevice(ctx context.Context, in RegisterDeviceInput) error {
	in.DeviceToken = strings.TrimSpace(in.DeviceToken)
	in.Platform = strings.ToLower(strings.TrimSpace(in.Platform))

	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	clm := pkgjwt.GetAuth[pkgjwt.AccessTokenPayload](ctx)

	if err := s.repoDB.NotificationUserDeviceRegister(ctx, clm.Payload().UserID, in.DeviceToken, in.Platform); err != nil {
		slog.ErrorContext(ctx, "failed to repo register device token", "user_id", clm.Payload().UserID, "error", err)
		return pkgerror.NewServer(err)
	}

	return nil
}

func (s *Usecase) RemoveDevice(ctx context.Context, in RemoveDeviceInput) error {
	in.DeviceToken = strings.TrimSpace(in.DeviceToken)

	if err := s.validator.Validate(in); err != nil {
		return pkgerror.NewInvalidInput(err)
	}

	if err := s.repoDB.NotificationUserDeviceRemove(ctx, in.DeviceToken); err != nil {
		slog.ErrorContext(ctx, "failed to repo remove device token", "error", err)
		return pkgerror.NewServer(err)
	}

	return nil
}
