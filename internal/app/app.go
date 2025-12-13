package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkglog"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type App struct {
	// configuration
	config pkgconfig.Config

	// libraries
	goroutine       *pkgroutine.Manager
	validator       pkgvalidator.Validator
	clock           pkgclock.Clocker
	hash            pkghash.Hash
	uid             pkguid.NumberID
	uuid            pkguid.StringID
	mail            pkgmail.Mail
	jwtTempToken    pkgjwt.JWT[map[string]any]
	jwtAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	jwtRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]

	// resources
	dbConn    *pgxpool.Pool
	cacheConn *redis.Client

	// server
	router     chi.Router
	httpServer *http.Server

	//
	closerFn map[string]func(context.Context) error
}

func New() *App {
	pkglog.InitLogging()

	app := &App{}

	app.initConfig()
	app.initLibraries()
	app.initJWT()
	app.initDatabase()
	app.initCache()
	app.initMail()
	app.initHTTPServer()
	app.initModules()
	app.initClosers()

	return app
}
