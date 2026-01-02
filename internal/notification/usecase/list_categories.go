package usecase

import (
	"context"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

func (s *Usecase) ListCategories(ctx context.Context) (_ []entity.Category, err error) {
	ctx, span := s.startSpan(ctx, "ListCategories")
	defer span.End()

	if _, err = s.requireAuth(ctx); err != nil {
		return nil, err
	}

	items, err := s.repoDB.ListCategories(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo list notification categories", "error", err)
		return nil, goerror.NewServer(err)
	}

	return items, nil
}
