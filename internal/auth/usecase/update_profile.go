package usecase

import (
	"context"
	"log/slog"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

type UpdateProfileInput struct {
	FullName string `validate:"required,min=2,max=100,alphaspace"`
}

func (s *Usecase) UpdateProfile(ctx context.Context, in UpdateProfileInput) error {
	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	userID := clm.GetInt64(keyPayloadUserID)
	fullName := strings.TrimSpace(in.FullName)

	if err := s.repoDB.UpdateUserProfile(ctx, userID, fullName); err != nil {
		slog.ErrorContext(ctx, "failed to update user profile", "user_id", userID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
