package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/shared/constant"
)

type UserListInput struct {
	Search    string // value already trimmed
	Status    entity.UserStatus
	Size      int32
	Page      int32
	SortBy    string // value already trimmed
	SortOrder string // value is: `asc` or `desc`; already trimmed and lowered
}

type UserListOutput struct {
	Page  int32
	Size  int32
	Total int64
	Users []entity.User
}

func (s *Usecase) UserList(ctx context.Context, in UserListInput) (_ *UserListOutput, err error) {
	ctx, span := s.startSpan(ctx, "ListUsers")
	defer span.End()

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	if err := s.isAuthorized(ctx, clm.Subject, constant.PermIdentityMgmtUsers, constant.PermActRead); err != nil {
		return nil, err
	}

	user, err := s.repoDB.GetUserByEmail(ctx, clm.UserEmail, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "email", clm.UserEmail)
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", clm.UserEmail, "error", err)
		return nil, goerror.NewServer(err)
	}

	if err := s.ensureUserStatusAllowed(ctx, user.ID, user.Status); err != nil {
		return nil, err
	}

	if in.Size <= 0 || in.Size > 100 {
		in.Size = 10 // default limit
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	filterData := entity.UserListFilterData{
		IsFilterBySearch: false,
		IsFilterByStatus: false,
		OrderBy:          in.SortBy,
		OrderDirection:   in.SortOrder,
		Search:           in.Search,
		Status:           in.Status,
		Size:             in.Size,
		Page:             (in.Page - 1) * in.Size,
	}
	if in.Status != entity.UserStatusUnknown {
		filterData.IsFilterByStatus = true
	}
	if in.Search != "" {
		filterData.IsFilterBySearch = true
	}

	users, count, err := s.repoDB.GetUserList(ctx, filterData)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo list users", "error", err)
		return nil, goerror.NewServer(err)
	}

	return &UserListOutput{
		Page:  in.Page,
		Size:  in.Size,
		Total: count,
		Users: users,
	}, nil
}
