package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
)

func (s *Usecase) RefreshToken(ctx context.Context, in domain.RefreshTokenInput) (*domain.RefreshTokenOutput, error) {
	return nil, nil
}
