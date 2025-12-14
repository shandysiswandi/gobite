package db

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func mfaFactorFromSQL(u pkgsql.MfaFactor) *domain.MfaFactor {
	item := &domain.MfaFactor{
		ID:           u.ID,
		UserID:       u.UserID,
		Type:         u.Type,
		FriendlyName: u.FriendlyName,
		Secret:       u.Secret,
		KeyVersion:   u.KeyVersion,
		IsVerified:   u.IsVerified,
	}

	if u.CreatedAt.Valid {
		item.CreatedAt = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		item.UpdatedAt = u.UpdatedAt.Time
	}

	return item
}

func (s *SQL) MfaFactorGetByUserID(ctx context.Context, userID int64) ([]domain.MfaFactor, error) {
	items, err := s.query.MfaFactorGetByUserID(ctx, pkgsql.MfaFactorGetByUserIDParams{
		UserID:     userID,
		IsVerified: true,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.MfaFactor, 0)
	for _, item := range items {
		result = append(result, *mfaFactorFromSQL(item))
	}

	return result, nil
}
