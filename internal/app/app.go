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
	"github.com/shandysiswandi/gobite/internal/pkg/pkglog"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type App struct {
	// configuration
	config pkgconfig.Config

	// libraries
	validator pkgvalidator.Validator
	clock     pkgclock.Clocker
	hash      pkghash.Hash
	uid       pkguid.NumberID
	uuid      pkguid.StringID

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
	app.initDatabase()
	app.initCache()
	app.initHTTPServer()
	app.initModules()
	app.initClosers()

	return app
}
