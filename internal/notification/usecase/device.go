package usecase

import (
	"context"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

func (s *Usecase) RegisterDevice(ctx context.Context, in RegisterDeviceInput) error {
	in.DeviceToken = strings.TrimSpace(in.DeviceToken)
	in.Platform = strings.ToLower(strings.TrimSpace(in.Platform))

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	uid := clm.GetInt64(keyPayloadUserID)

	if err := s.repoDB.NotificationUserDeviceRegister(ctx, uid, in.DeviceToken, in.Platform); err != nil {
		slog.ErrorContext(ctx, "failed to repo register device token", "user_id", uid, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}

func (s *Usecase) RemoveDevice(ctx context.Context, in RemoveDeviceInput) error {
	in.DeviceToken = strings.TrimSpace(in.DeviceToken)

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	if err := s.repoDB.NotificationUserDeviceRemove(ctx, in.DeviceToken); err != nil {
		slog.ErrorContext(ctx, "failed to repo remove device token", "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
