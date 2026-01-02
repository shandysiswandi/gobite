package app

import (
	"log/slog"
	"os"

	"github.com/shandysiswandi/gobite/internal/identity"
	"github.com/shandysiswandi/gobite/internal/notification"
)

func (a *App) initModules() {
	if a.config.GetBool("modules.identity.enabled") {
		if err := identity.New(identity.Dependency{
			Config:          a.config,
			Instrument:      a.ins,
			UID:             a.uid,
			UUID:            a.uuid,
			OID:             a.oid,
			Bcrypt:          a.bcrypt,
			HMAC:            a.hmac,
			Argon2ID:        a.argon2id,
			MFAEncryptor:    a.mfaEncryptor,
			MFARecoveryCode: a.mfaRecoveryCode,
			Clock:           a.clock,
			Validator:       a.validator,
			Router:          a.router,
			Totp:            a.totp,
			DBConn:          a.dbConn,
			CacheConn:       a.cacheConn,
			Idempotency:     a.idemp,
			Messaging:       a.messaging,
			Storage:         a.storage,
			Goroutine:       a.goroutine,
			JWT:             a.jwt,
			Enforcer:        a.casbin,
		}); err != nil {
			slog.Error("failed to init module identity", "error", err)
			os.Exit(1)
		}
	}

	if a.config.GetBool("modules.notification.enabled") {
		if err := notification.New(notification.Dependency{
			Ctx:        a.ctx,
			DBConn:     a.dbConn,
			Messaging:  a.messaging,
			Config:     a.config,
			Instrument: a.ins,
			UID:        a.uid,
			UUID:       a.uuid,
			Clock:      a.clock,
			Goroutine:  a.goroutine,
			Validator:  a.validator,
			Router:     a.router,
			Mail:       a.mail,
			JWT:        a.jwt,
		}); err != nil {
			slog.Error("failed to init module notification", "error", err)
			os.Exit(1)
		}
	}
}
