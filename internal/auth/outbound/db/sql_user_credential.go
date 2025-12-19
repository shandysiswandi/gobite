package db

// import (
// 	"context"
// 	"errors"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/shandysiswandi/gobite/internal/auth/entity"
// 	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
// 	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
// )

// func userCredentialFromSQL(u sqlc.AuthUserCredential) *entity.UserCredential {
// 	item := &entity.UserCredential{
// 		UserID:   u.UserID,
// 		Password: u.Password,
// 	}

// 	if u.UpdatedAt.Valid {
// 		item.UpdatedAt = u.UpdatedAt.Time
// 	}

// 	return item
// }

// func (s *DB) UserCredentialGetByUserID(ctx context.Context, userID int64) (*entity.UserCredential, error) {
// 	result, err := s.query.UserCredentialGetByUserID(ctx, userID)
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return nil, goerror.ErrNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	return userCredentialFromSQL(result), nil
// }
