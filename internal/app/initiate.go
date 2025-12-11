package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

func (a *App) initConfig() {
	path := "/config/config.yaml"
	if os.Getenv("LOCAL") == "true" {
		path = "./config/config.yaml"
	}

	cfg, err := pkgconfig.NewViper(path)
	if err != nil {
		slog.Error("failed to init config", "error", err)
		os.Exit(1)
	}

	//nolint:errcheck,gosec // ignore error
	os.Setenv("TZ", cfg.GetString("tz"))

	a.config = cfg
}

func (a *App) initLibraries() {
	validator, err := pkgvalidator.NewV10Validator()
	if err != nil {
		slog.Error("failed to init validation v10 validator", "error", err)
		os.Exit(1)
	}

	snow, err := pkguid.NewSnowflake()
	if err != nil {
		slog.Error("failed to init uid number snowflake", "error", err)
		os.Exit(1)
	}

	a.uid = snow
	a.validator = validator
	a.clock = pkgclock.New()
	a.uuid = pkguid.NewUUID()
	a.hash = pkghash.NewBcrypt(int(a.config.GetInt("password.cost")), a.config.GetString("password.secret"))
}

func (a *App) initDatabase() {
	config, err := pgxpool.ParseConfig(a.config.GetString("database.dsn"))
	if err != nil {
		slog.Error("failed to parse DB connection string.", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		slog.Error("failed to create DB connection pool", "error", err)
		os.Exit(1)
	}

	a.dbConn = pool
}

func (a *App) initCache() {
	rdb := redis.NewClient(&redis.Options{
		Addr: a.config.GetString("redis.addr"),
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to init redis", "error", err)
		os.Exit(1)
	}

	a.cacheConn = rdb
}

func (a *App) initHTTPServer() {
	a.router = pkgrouter.NewChi(a.uuid)

	a.httpServer = &http.Server{
		Addr:              a.config.GetString("server.address.http"),
		Handler:           a.router,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

//nolint:unparam // is always nil
func (a *App) initClosers() {
	a.closerFn = map[string]func(context.Context) error{
		"HTTP Server": func(ctx context.Context) error {
			return a.httpServer.Shutdown(ctx)
		},
		"Database": func(context.Context) error {
			a.dbConn.Close()

			return nil
		},
		"Redis": func(context.Context) error {
			return a.cacheConn.Close()
		},
		"Config": func(context.Context) error {
			return a.config.Close()
		},
	}
}
