package usecase

import (
	"context"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type (
	DeviceRegisterInput struct {
		DeviceToken string `validate:"required"`
		Platform    string `validate:"required,oneof=android ios web"`
	}
)

func (s *Usecase) DeviceRegister(ctx context.Context, in DeviceRegisterInput) error {
	ctx, span := s.startSpan(ctx, "DeviceRegister")
	defer span.End()

	in.DeviceToken = strings.TrimSpace(in.DeviceToken)
	in.Platform = strings.ToLower(strings.TrimSpace(in.Platform))

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	if err := s.repoDB.RegisterUserDevice(ctx, clm.UserID, in.DeviceToken, in.Platform); err != nil {
		slog.ErrorContext(ctx, "failed to repo register device token", "user_id", clm.UserID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
