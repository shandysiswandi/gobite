package app

import (
	"context"
	"net/http"

	"github.com/casbin/casbin/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/hash"
	"github.com/shandysiswandi/gobite/internal/pkg/idempotency"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/pkg/mfa"
	"github.com/shandysiswandi/gobite/internal/pkg/otp"
	"github.com/shandysiswandi/gobite/internal/pkg/pgxcasbin"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
	"github.com/shandysiswandi/gobite/internal/pkg/storage"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
)

// App wires dependencies and manages service lifecycle.
type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	// configuration
	config config.Config
	ins    instrument.Instrumentation

	// libraries
	goroutine       *goroutine.Manager
	validator       validator.Validator
	clock           clock.Clocker
	hmac            hash.Hash
	argon2id        hash.Hash
	bcrypt          hash.Hash
	uid             uid.NumberID
	oid             uid.StringID
	uuid            uid.StringID
	totp            otp.OTP
	jwt             jwt.JWT
	mfaEncryptor    mfa.Encryptor
	mfaRecoveryCode mfa.RecoveryCodeGenerator

	// resources
	dbConn        *pgxpool.Pool
	cacheConn     *redis.Client
	idemp         idempotency.Idempotency
	mail          mail.Mail
	messaging     messaging.Messaging
	storage       storage.Storage
	casbin        *casbin.Enforcer
	casbinWatcher *pgxcasbin.Watcher

	// server
	router     *router.Router
	httpServer *http.Server
	sseServer  *http.Server

	//
	closers []struct {
		name string
		fn   func(context.Context) error
	}
}

// New initializes the application with default wiring and returns an App instance.
func New() *App {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		ctx:    ctx,
		cancel: cancel,
	}

	app.initConfig()
	app.initInstrument()
	app.initLibraries()
	app.initJWT()
	app.initDatabase()
	app.initCache()
	app.initMail()
	app.initStorage()
	app.initMessaging()
	app.initCasbin()
	app.initHTTPServer()
	app.initModules()
	app.initClosers()

	return app
}
