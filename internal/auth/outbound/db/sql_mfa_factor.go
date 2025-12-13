package db

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgsql"
)

func mfaFactorFromSQL(u pkgsql.MfaFactor) *domain.MfaFactor {
	return &domain.MfaFactor{
		ID:           u.ID,
		UserID:       u.UserID,
		Type:         u.Type,
		FriendlyName: u.FriendlyName,
		Secret:       u.Secret,
		KeyVersion:   u.KeyVersion,
		IsVerified:   u.IsVerified,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
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
