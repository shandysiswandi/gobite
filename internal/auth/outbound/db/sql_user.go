package db

// import (
// 	"context"
// 	"errors"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/shandysiswandi/gobite/internal/auth/entity"
// 	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
// 	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
// )

// func userFromSQL(u sqlc.AuthUser, err error) (*entity.User, error) {
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return nil, goerror.ErrNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	item := &entity.User{
// 		ID:        u.ID,
// 		Email:     u.Email,
// 		FullName:  u.FullName,
// 		AvatarURL: u.AvatarUrl,
// 		Status:    u.Status,
// 	}

// 	if u.CreatedAt.Valid {
// 		item.CreatedAt = u.CreatedAt.Time
// 	}
// 	if u.UpdatedAt.Valid {
// 		item.UpdatedAt = u.UpdatedAt.Time
// 	}

// 	return item, nil
// }

// func (s *DB) UserGetByEmailStatus(ctx context.Context, email string, sts entity.UserStatus) (*entity.User, error) {
// 	return userFromSQL(s.query.UserGetByEmailStatus(ctx, sqlc.UserGetByEmailStatusParams{
// 		Email:  email,
// 		Status: sts,
// 	}))
// }

// func (s *DB) UserGetByID(ctx context.Context, id int64) (*entity.User, error) {
// 	return userFromSQL(s.query.UserGetByID(ctx, id))
// }

// func (s *DB) UserGetByIDStatus(ctx context.Context, id int64, sts entity.UserStatus) (*entity.User, error) {
// 	return userFromSQL(s.query.UserGetByIDStatus(ctx, sqlc.UserGetByIDStatusParams{
// 		ID:     id,
// 		Status: sts,
// 	}))
// }

// func (s *DB) UserUpdateStatus(ctx context.Context, id int64, oldStatus, newStatus entity.UserStatus) error {
// 	return s.query.UserUpdateStatus(ctx, sqlc.UserUpdateStatusParams{
// 		ID:        id,
// 		NewStatus: newStatus,
// 		OldStatus: oldStatus,
// 	})
// }
