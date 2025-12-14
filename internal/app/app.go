package app

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkglog"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgotp"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	// configuration
	config pkgconfig.Config

	// libraries
	goroutine       *pkgroutine.Manager
	validator       pkgvalidator.Validator
	clock           pkgclock.Clocker
	hash            pkghash.Hash
	uid             pkguid.NumberID
	uuid            pkguid.StringID
	totp            pkgotp.OTP
	jwtTempToken    pkgjwt.JWT[map[string]any]
	jwtAccessToken  pkgjwt.JWT[pkgjwt.AccessTokenPayload]
	jwtRefreshToken pkgjwt.JWT[pkgjwt.RefreshTokenPayload]

	// resources
	dbConn    *pgxpool.Pool
	cacheConn *redis.Client
	mail      pkgmail.Mail
	messaging pkgmessaging.Messaging

	// server
	router     *pkgrouter.Router
	httpServer *http.Server

	//
	closerFn map[string]func(context.Context) error
}

func New() *App {
	pkglog.InitLogging()

	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		ctx:    ctx,
		cancel: cancel,
	}

	app.initConfig()
	app.initLibraries()
	app.initJWT()
	app.initDatabase()
	app.initCache()
	app.initMail()
	app.initMessaging()
	app.initHTTPServer()
	app.initModules()
	app.initClosers()

	return app
}
