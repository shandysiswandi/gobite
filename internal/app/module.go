package app

import (
	"log/slog"
	"os"

	"github.com/shandysiswandi/gobite/internal/auth"
)

func (a *App) initModules() {
	if a.config.GetBool("modules.auth.enabled") {
		if err := auth.New(auth.Dependency{
			Config:          a.config,
			UID:             a.uid,
			UUID:            a.uuid,
			Hash:            a.hash,
			Clock:           a.clock,
			Validator:       a.validator,
			Router:          a.router,
			Totp:            a.totp,
			DBConn:          a.dbConn,
			CacheConn:       a.cacheConn,
			Messaging:       a.messaging,
			Goroutine:       a.goroutine,
			JWTTempToken:    a.jwtTempToken,
			JWTAccessToken:  a.jwtAccessToken,
			JWTRefreshToken: a.jwtRefreshToken,
		}); err != nil {
			slog.Error("failed to init module auth", "error", err)
			os.Exit(1)
		}
	}

	//
}
