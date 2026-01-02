package usecase

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
)

func (s *Usecase) requireAuth(ctx context.Context) (*jwt.Claims, error) {
	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return nil, goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	return clm, nil
}
