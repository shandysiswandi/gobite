package app

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/logging"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/pkg/mfacrypto"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	// configuration
	config config.Config

	// libraries
	goroutine *goroutine.Manager
	validator validator.Validator
	clock     clock.Clocker
	hash      hash.Hash
	password  hash.Hash
	uid       uid.NumberID
	oid       uid.StringID
	uuid      uid.StringID
	totp      otp.OTP
	jwt       jwt.JWT
	mfacry    mfacrypto.Encryptor

	// resources
	dbConn    *pgxpool.Pool
	cacheConn *redis.Client
	mail      mail.Mail
	messaging messaging.Messaging

	// server
	router     *router.Router
	httpServer *http.Server

	//
	closerFn map[string]func(context.Context) error
}

func New() *App {
	logging.InitLogging()

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
