package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/shared/constant"
)

type UserListInput struct {
	Search    string // value already trimmed
	Statuses  []string
	DateFrom  time.Time
	DateTo    time.Time
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

func (s *Usecase) UserList(ctx context.Context, in UserListInput) (*UserListOutput, error) {
	ctx, span := s.startSpan(ctx, "ListUsers")
	defer span.End()

	if _, err := s.authenticatedAndAuthorized(ctx, constant.PermIdentityMgmtUsers, constant.PermActCreate); err != nil {
		return nil, err
	}

	if in.Size <= 0 || in.Size > 100 {
		in.Size = 10 // default limit
	}
	filterData := entity.UserListFilterData{
		OrderBy:        in.SortBy,
		OrderDirection: in.SortOrder,
		Search:         in.Search,
		Statuses:       entity.ToInt16Slice(entity.ParseSafeUserStatuses(in.Statuses)),
		DateFrom:       in.DateFrom,
		DateTo:         in.DateTo,
		Size:           in.Size,
		Page:           (max(in.Page, 1) - 1) * in.Size,
	}
	if in.Search != "" {
		filterData.IsFilterBySearch = true
	}
	if len(filterData.Statuses) > 0 {
		filterData.IsFilterByStatus = true
	}

	users, count, err := s.repoDB.GetUserList(ctx, filterData)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo list users", "error", err)
		return nil, goerror.NewServer(err)
	}

	return &UserListOutput{
		Page:  max(in.Page, 1),
		Size:  in.Size,
		Total: count,
		Users: users,
	}, nil
}
