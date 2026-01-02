package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/shared/constant"
)

type (
	UserDetailInput struct {
		ID int64 `validate:"required,gt=0"`
	}

	UserDetailOutput struct {
		User entity.User
	}
)

func (s *Usecase) UserDetail(ctx context.Context, in UserDetailInput) (*UserDetailOutput, error) {
	ctx, span := s.startSpan(ctx, "UserDetail")
	defer span.End()

	_, err := s.authenticatedAndAuthorized(ctx, constant.PermIdentityMgmtUsers, constant.PermActCreate)
	if err != nil {
		return nil, err
	}

	user, err := s.repoDB.GetUserByID(ctx, in.ID, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user not found", "user_id", in.ID)
		return nil, goerror.NewBusiness("user not found", goerror.CodeNotFound)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by id", "user_id", in.ID, "error", err)
		return nil, goerror.NewServer(err)
	}

	return &UserDetailOutput{User: *user}, nil
}
