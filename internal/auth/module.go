package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/auth/inbound"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/cache"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/db"
	"github.com/shandysiswandi/gobite/internal/auth/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type Dependency struct {
	DBConn    *pgxpool.Pool
	CacheConn *redis.Client
	Config    pkgconfig.Config
	UID       pkguid.NumberID
	Hash      pkghash.Hash
	Clock     pkgclock.Clocker
	Validator pkgvalidator.Validator
	Router    chi.Router
}

func New(dep Dependency) error {
	dbAuth := db.NewSQL(dep.DBConn)
	cacheAuth := cache.NewRedis(dep.CacheConn)

	uc := usecase.NewAuth(usecase.Dependency{
		RepoDB:    dbAuth,
		RepoCache: cacheAuth,
		Validator: dep.Validator,
		Hash:      dep.Hash,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)

	return nil
}
