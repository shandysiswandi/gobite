package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/pkg/router"
)

type uc interface {
	MePermissions(ctx context.Context) (map[string][]string, error)
}

func RegisterHTTPEndpoint(r *router.Router, uc uc) {
	end := &HTTPEndpoint{uc: uc}

	r.GET("/api/v1/iam/me/permissions", end.MePermissions)
}
