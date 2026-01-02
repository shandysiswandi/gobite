package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/shared/constant"
)

type (
	UserDeleteInput struct {
		ID int64 `validate:"required,gt=0"`
	}
)

func (s *Usecase) UserDelete(ctx context.Context, in UserDeleteInput) error {
	ctx, span := s.startSpan(ctx, "UserDelete")
	defer span.End()

	if err := s.validator.Validate(in); err != nil {
		return goerror.NewInvalidInput(err)
	}

	clm, err := s.authenticatedAndAuthorized(ctx, constant.PermIdentityMgmtUsers, constant.PermActCreate)
	if err != nil {
		return err
	}

	user, err := s.repoDB.GetUserByID(ctx, in.ID, true)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user not found", "user_id", in.ID)
		return goerror.NewBusiness("user not found", goerror.CodeNotFound)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to get user by id", "user_id", in.ID, "error", err)
		return goerror.NewServer(err)
	}

	if user.DeletedAt != nil {
		return nil
	}

	if err := s.repoDB.MarkUserDeleted(ctx, user.ID, clm.UserID); err != nil {
		slog.ErrorContext(ctx, "failed to mark user deleted", "user_id", user.ID, "by_user_id", clm.UserID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}
