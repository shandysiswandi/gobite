package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/auth/inbound"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/cache"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/db"
	"github.com/shandysiswandi/gobite/internal/auth/outbound/mq"
	"github.com/shandysiswandi/gobite/internal/auth/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/pkg/mfacrypto"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
)

type Dependency struct {
	DBConn    *pgxpool.Pool
	CacheConn *redis.Client
	Messaging messaging.Messaging
	Config    config.Config
	UID       uid.NumberID
	UUID      uid.StringID
	OID       uid.StringID
	Password  hash.Hash
	Hash      hash.Hash
	MFACrypto mfacrypto.Encryptor
	Clock     clock.Clocker
	Totp      otp.OTP
	Goroutine *goroutine.Manager
	Validator validator.Validator
	JWT       jwt.JWT
	Router    *router.Router
}

func New(dep Dependency) error {
	dbAuth := db.NewDB(dep.DBConn)
	cacheAuth := cache.NewRedis(dep.CacheConn, dep.Config)
	repoMsg := mq.NewMessaging(dep.Messaging)

	uc := usecase.NewAuth(usecase.Dependency{
		RepoDB:        dbAuth,
		RepoCache:     cacheAuth,
		RepoMessaging: repoMsg,
		Validator:     dep.Validator,
		Config:        dep.Config,
		Password:      dep.Password,
		Hash:          dep.Hash,
		MFACrypto:     dep.MFACrypto,
		UID:           dep.UID,
		UUID:          dep.UUID,
		OID:           dep.OID,
		Totp:          dep.Totp,
		Clock:         dep.Clock,
		JWT:           dep.JWT,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)

	return nil
}
