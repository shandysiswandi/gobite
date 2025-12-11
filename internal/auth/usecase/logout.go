package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/auth/domain"
)

func (s *Usecase) Logout(ctx context.Context, in domain.LogoutInput) (*domain.LogoutOutput, error) {
	return nil, nil
}
