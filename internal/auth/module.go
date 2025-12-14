package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/auth/inbound"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/cache"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/db"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/mq"
	"github.com/shandysiswandi/gobite/internal/auth/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgotp"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type Dependency struct {
	DBConn          *pgxpool.Pool
	CacheConn       *redis.Client
	Messaging       pkgmessaging.Messaging
	Config          pkgconfig.Config
	UID             pkguid.NumberID
	UUID            pkguid.StringID
	Hash            pkghash.Hash
	Clock           pkgclock.Clocker
	Totp            pkgotp.OTP
	Goroutine       *pkgroutine.Manager
	Validator       pkgvalidator.Validator
	JWTTempToken    pkgjwt.JWT[map[string]any]
	JWTAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	JWTRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]
	Router          *pkgrouter.Router
}

func New(dep Dependency) error {
	dbAuth := db.NewSQL(dep.DBConn)
	cacheAuth := cache.NewRedis(dep.CacheConn, dep.Config)
	repoMsg := mq.NewMessaging(dep.Messaging)

	uc := usecase.NewAuth(usecase.Dependency{
		RepoDB:          dbAuth,
		RepoCache:       cacheAuth,
		RepoMessaging:   repoMsg,
		Validator:       dep.Validator,
		Config:          dep.Config,
		Hash:            dep.Hash,
		UID:             dep.UID,
		UUID:            dep.UUID,
		Totp:            dep.Totp,
		Clock:           dep.Clock,
		JWTTempToken:    dep.JWTTempToken,
		JWTAccessToken:  dep.JWTAccessToken,
		JWTRefreshToken: dep.JWTRefreshToken,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)

	return nil
}
