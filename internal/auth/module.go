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
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgotp"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type Dependency struct {
	DBConn          *pgxpool.Pool
	CacheConn       *redis.Client
	Mail            pkgmail.Mail
	Config          pkgconfig.Config
	UID             pkguid.NumberID
	UUID            pkguid.StringID
	Hash            pkghash.Hash
	Clock           pkgclock.Clocker
	Totp            pkgotp.OTP
	Validator       pkgvalidator.Validator
	JWTTempToken    pkgjwt.JWT[map[string]any]
	JWTAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	JWTRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]
	Router          chi.Router
}

func New(dep Dependency) error {
	dbAuth := db.NewSQL(dep.DBConn)
	cacheAuth := cache.NewRedis(dep.CacheConn, dep.Config)

	uc := usecase.NewAuth(usecase.Dependency{
		RepoDB:          dbAuth,
		RepoCache:       cacheAuth,
		Mail:            dep.Mail,
		Validator:       dep.Validator,
		Hash:            dep.Hash,
		UID:             dep.UID,
		Totp:            dep.Totp,
		Clock:           dep.Clock,
		JWTTempToken:    dep.JWTTempToken,
		JWTAccessToken:  dep.JWTAccessToken,
		JWTRefreshToken: dep.JWTRefreshToken,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)

	return nil
}
