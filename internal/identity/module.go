package identity

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/identity/inbound"
	"github.com/shandysiswandi/gobite/internal/identity/outbound/db"
	"github.com/shandysiswandi/gobite/internal/identity/outbound/mq"
	"github.com/shandysiswandi/gobite/internal/identity/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/enforcer"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
	"github.com/shandysiswandi/gobite/internal/pkg/idempotency"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/pkg/mfa"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
	"github.com/shandysiswandi/gobite/internal/pkg/storage"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
)

type Dependency struct {
	DBConn          *pgxpool.Pool
	CacheConn       *redis.Client
	Idempotency     idempotency.Idempotency
	Messaging       messaging.Messaging
	Storage         storage.Storage
	Config          config.Config
	Instrument      instrument.Instrumentation
	UID             uid.NumberID
	UUID            uid.StringID
	OID             uid.StringID
	HMAC            hash.Hash
	Bcrypt          hash.Hash
	Argon2ID        hash.Hash
	MFAEncryptor    mfa.Encryptor
	MFARecoveryCode mfa.RecoveryCodeGenerator
	Clock           clock.Clocker
	Totp            otp.OTP
	Goroutine       *goroutine.Manager
	Validator       validator.Validator
	JWT             jwt.JWT
	Enforcer        enforcer.Enforcer
	Router          *router.Router
}

func New(dep Dependency) error {
	dbAuth := db.NewDB(dep.DBConn, dep.Instrument)
	repoMsg := mq.NewMessaging(dep.Messaging, dep.Instrument)

	uc := usecase.New(usecase.Dependency{
		RepoDB:          dbAuth,
		RepoMessaging:   repoMsg,
		Idempotency:     dep.Idempotency,
		Validator:       dep.Validator,
		Config:          dep.Config,
		Storage:         dep.Storage,
		HMAC:            dep.HMAC,
		Bcrypt:          dep.Bcrypt,
		Argon2ID:        dep.Argon2ID,
		MFAEncryptor:    dep.MFAEncryptor,
		MFARecoveryCode: dep.MFARecoveryCode,
		UID:             dep.UID,
		UUID:            dep.UUID,
		OID:             dep.OID,
		Totp:            dep.Totp,
		Clock:           dep.Clock,
		JWT:             dep.JWT,
		Instrument:      dep.Instrument,
		Enforcer:        dep.Enforcer,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)

	return nil
}
