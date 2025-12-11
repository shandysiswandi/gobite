package app

import (
	"log/slog"
	"os"

	"github.com/shandysiswandi/gobite/internal/auth"
)

func (a *App) initModules() {
	if a.config.GetBool("modules.auth.enabled") {
		if err := auth.New(auth.Dependency{
			Config:    a.config,
			UID:       a.uid,
			Hash:      a.hash,
			Clock:     a.clock,
			Validator: a.validator,
			Router:    a.router,
		}); err != nil {
			slog.Error("failed to init module auth", "error", err)
			os.Exit(1)
		}
	}

	//
}
