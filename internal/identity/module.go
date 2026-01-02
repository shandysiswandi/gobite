package identity

import (
	"github.com/casbin/casbin/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/identity/inbound"
	"github.com/shandysiswandi/gobite/internal/identity/outbound/db"
	"github.com/shandysiswandi/gobite/internal/identity/outbound/mq"
	"github.com/shandysiswandi/gobite/internal/identity/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
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
	DBConn          *pgxpool.Pool              `validate:"required"`
	CacheConn       *redis.Client              `validate:"required"`
	Goroutine       *goroutine.Manager         `validate:"required"`
	Enforcer        *casbin.Enforcer           `validate:"required"`
	Router          *router.Router             `validate:"required"`
	Idempotency     idempotency.Idempotency    `validate:"required"`
	Messaging       messaging.Messaging        `validate:"required"`
	Storage         storage.Storage            `validate:"required"`
	Config          config.Config              `validate:"required"`
	Instrument      instrument.Instrumentation `validate:"required"`
	UID             uid.NumberID               `validate:"required"`
	UUID            uid.StringID               `validate:"required"`
	OID             uid.StringID               `validate:"required"`
	HMAC            hash.Hash                  `validate:"required"`
	Bcrypt          hash.Hash                  `validate:"required"`
	Argon2ID        hash.Hash                  `validate:"required"`
	MFAEncryptor    mfa.Encryptor              `validate:"required"`
	MFARecoveryCode mfa.RecoveryCodeGenerator  `validate:"required"`
	Clock           clock.Clocker              `validate:"required"`
	Totp            otp.OTP                    `validate:"required"`
	Validator       validator.Validator        `validate:"required"`
	JWT             jwt.JWT                    `validate:"required"`
}

func New(dep Dependency) error {
	if err := dep.Validator.Validate(dep); err != nil {
		return err
	}

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
		Goroutine:       dep.Goroutine,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)

	return nil
}
