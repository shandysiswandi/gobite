package app

import (
	"log/slog"
	"os"

	"github.com/shandysiswandi/gobite/internal/auth"
	"github.com/shandysiswandi/gobite/internal/notification"
)

func (a *App) initModules() {
	if a.config.GetBool("modules.auth.enabled") {
		if err := auth.New(auth.Dependency{
			Config:    a.config,
			UID:       a.uid,
			UUID:      a.uuid,
			OID:       a.oid,
			Password:  a.password,
			Hash:      a.hash,
			MFACrypto: a.mfacry,
			Clock:     a.clock,
			Validator: a.validator,
			Router:    a.router,
			Totp:      a.totp,
			DBConn:    a.dbConn,
			CacheConn: a.cacheConn,
			Messaging: a.messaging,
			Goroutine: a.goroutine,
			JWT:       a.jwt,
		}); err != nil {
			slog.Error("failed to init module auth", "error", err)
			os.Exit(1)
		}
	}

	if a.config.GetBool("modules.notification.enabled") {
		if err := notification.New(notification.Dependency{
			Ctx:       a.ctx,
			DBConn:    a.dbConn,
			Messaging: a.messaging,
			Config:    a.config,
			UID:       a.uid,
			Clock:     a.clock,
			Goroutine: a.goroutine,
			Validator: a.validator,
			Router:    a.router,
			Mail:      a.mail,
			JWT:       a.jwt,
		}); err != nil {
			slog.Error("failed to init module notification", "error", err)
			os.Exit(1)
		}
	}
}
