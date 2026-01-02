package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/shandysiswandi/gobite/internal/identity/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/shared/constant"
)

const userExportPageSize int32 = 1_000

type (
	UserExportInput struct {
		Search    string
		Statuses  []string
		DateFrom  time.Time
		DateTo    time.Time
		SortBy    string
		SortOrder string
	}

	UserExportOutput struct {
		Users []entity.User
	}
)

func (s *Usecase) UserExport(ctx context.Context, in UserExportInput) (*UserExportOutput, error) {
	ctx, span := s.startSpan(ctx, "UserExport")
	defer span.End()

	_, err := s.authenticatedAndAuthorized(ctx, constant.PermIdentityMgmtUsers, constant.PermActCreate)
	if err != nil {
		return nil, err
	}

	filterData := entity.UserListFilterData{
		OrderBy:        in.SortBy,
		OrderDirection: in.SortOrder,
		Search:         in.Search,
		Statuses:       entity.ToInt16Slice(entity.ParseSafeUserStatuses(in.Statuses)),
		DateFrom:       in.DateFrom,
		DateTo:         in.DateTo,
		Size:           userExportPageSize,
		Page:           0,
	}
	if in.Search != "" {
		filterData.IsFilterBySearch = true
	}
	if len(filterData.Statuses) > 0 {
		filterData.IsFilterByStatus = true
	}

	var (
		users []entity.User
		page  int32 = 1
		total int64
	)

	for {
		filterData.Page = (page - 1) * userExportPageSize

		pageUsers, count, err := s.repoDB.GetUserList(ctx, filterData)
		if err != nil {
			slog.ErrorContext(ctx, "failed to repo export users", "error", err)
			return nil, goerror.NewServer(err)
		}

		if page == 1 {
			total = count
			if total == 0 {
				break
			}
			users = make([]entity.User, 0, min(total, int64(userExportPageSize)))
		}

		users = append(users, pageUsers...)

		if int64(len(users)) >= total || len(pageUsers) == 0 {
			break
		}

		page++
	}

	return &UserExportOutput{Users: users}, nil
}
