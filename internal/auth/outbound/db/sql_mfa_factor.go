package db

// import (
// 	"context"
// 	"errors"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/shandysiswandi/gobite/internal/auth/entity"
// 	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
// 	"github.com/shandysiswandi/gobite/internal/pkg/sqlc"
// )

// func mfaFactorFromSQL(u sqlc.AuthMfaFactor) *entity.MfaFactor {
// 	item := &entity.MfaFactor{
// 		ID:           u.ID,
// 		UserID:       u.UserID,
// 		Type:         u.Type,
// 		FriendlyName: u.FriendlyName,
// 		Secret:       u.Secret,
// 		KeyVersion:   u.KeyVersion,
// 		IsVerified:   u.IsVerified,
// 	}

// 	if u.CreatedAt.Valid {
// 		item.CreatedAt = u.CreatedAt.Time
// 	}
// 	if u.UpdatedAt.Valid {
// 		item.UpdatedAt = u.UpdatedAt.Time
// 	}

// 	return item
// }

// func (s *DB) MfaFactorGetByID(ctx context.Context, id int64, userID int64) (*entity.MfaFactor, error) {
// 	item, err := s.query.MfaFactorGetByID(ctx, sqlc.MfaFactorGetByIDParams{
// 		ID:     id,
// 		UserID: userID,
// 	})
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return nil, goerror.ErrNotFound
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	return mfaFactorFromSQL(item), nil
// }

// func (s *DB) MfaFactorVerify(ctx context.Context, id int64, userID int64) error {
// 	return s.query.MfaFactorVerify(ctx, sqlc.MfaFactorVerifyParams{
// 		ID:     id,
// 		UserID: userID,
// 	})
// }
