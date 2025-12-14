package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nsqio/go-nsq"
	"github.com/pquerna/otp"
	"github.com/redis/go-redis/v9"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkghash"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgotp"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgroutine"
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
	a.clock = pkgclock.New()
	a.uuid = pkguid.NewUUID()
	a.goroutine = pkgroutine.NewManager(100)
	a.hash = pkghash.NewBcrypt(int(a.config.GetInt("password.cost")), a.config.GetString("password.secret"))

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

	a.totp = pkgotp.NewTOTP(
		a.config.GetString("totp.issuer"),
		uint(a.config.GetInt("totp.period")),
		uint(a.config.GetInt("totp.skew")),
		otp.DigitsSix,
	)

	a.uid = snow
	a.validator = validator
}

func (a *App) initJWT() {
	acToken, err := pkgjwt.NewHS512[pkgjwt.AccessTokenPayload](pkgjwt.Config{
		Secret:   []byte(a.config.GetString("jwt.access.secret")),
		Issuer:   "gobite",
		Audience: "access",
		TTL:      time.Duration(a.config.GetInt("jwt.access.ttl")) * time.Minute,
		Clock:    a.clock,
		UUID:     a.uuid,
	})
	if err != nil {
		slog.Error("failed to init jwt access token", "error", err)
		os.Exit(1)
	}

	refToken, err := pkgjwt.NewHS512[pkgjwt.RefreshTokenPayload](pkgjwt.Config{
		Secret:   []byte(a.config.GetString("jwt.refresh.secret")),
		Issuer:   "gobite",
		Audience: "refresh",
		TTL:      time.Duration(a.config.GetInt("jwt.refresh.ttl")) * 24 * time.Hour,
		Clock:    a.clock,
		UUID:     a.uuid,
	})
	if err != nil {
		slog.Error("failed to init jwt refresh token", "error", err)
		os.Exit(1)
	}

	tempToken, err := pkgjwt.NewHS512[map[string]any](pkgjwt.Config{
		Secret:   []byte(a.config.GetString("jwt.temp.secret")),
		Issuer:   "gobite",
		Audience: "temp",
		TTL:      time.Duration(a.config.GetInt("jwt.temp.ttl")) * time.Minute,
		Clock:    a.clock,
		UUID:     a.uuid,
	})
	if err != nil {
		slog.Error("failed to init jwt temp token", "error", err)
		os.Exit(1)
	}

	a.jwtTempToken = tempToken
	a.jwtAccessToken = acToken
	a.jwtRefreshToken = refToken
}

func (a *App) initDatabase() {
	config, err := pgxpool.ParseConfig(a.config.GetString("database.url"))
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
	opt, err := redis.ParseURL(a.config.GetString("redis.url"))
	if err != nil {
		slog.Error("failed to parse redis url", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(opt)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to init redis", "error", err)
		os.Exit(1)
	}

	a.cacheConn = rdb
}

func (a *App) initMail() {
	mail, err := pkgmail.NewSMTP(pkgmail.SMTPConfig{
		Host:     a.config.GetString("mail.host"),
		Port:     int(a.config.GetInt("mail.port")),
		Username: a.config.GetString("mail.username"),
		Password: a.config.GetString("mail.password"),
		From:     a.config.GetString("mail.from"),
	})
	if err != nil {
		slog.Error("failed to init mail", "error", err)
		os.Exit(1)
	}

	a.mail = mail
}

func (a *App) initMessaging() {
	// consumer config
	ccfg := nsq.NewConfig()
	ccfg.MaxInFlight = int(a.config.GetInt("messaging.nsq.consumer_config.max_in_flight"))
	ccfg.LookupdPollInterval = time.Duration(a.config.GetInt("messaging.nsq.consumer_config.lookupd_poll_interval")) * time.Second

	// producer config
	pcfg := nsq.NewConfig()
	pcfg.MaxInFlight = int(a.config.GetInt("messaging.nsq.producer_config.max_in_flight"))
	pcfg.DialTimeout = time.Duration(a.config.GetInt("messaging.nsq.producer_config.dial_timeout")) * time.Second
	pcfg.ReadTimeout = time.Duration(a.config.GetInt("messaging.nsq.producer_config.read_timeout")) * time.Second
	pcfg.WriteTimeout = time.Duration(a.config.GetInt("messaging.nsq.producer_config.write_timeout")) * time.Second

	nsq, err := pkgmessaging.NewNSQ(pkgmessaging.NSQConfig{
		ProducerAddr:        a.config.GetString("messaging.nsq.producer_addr"),
		ConsumerNSQDAddrs:   a.config.GetArray("messaging.nsq.consumer_nsqd_addrs"),
		ConsumerLookupdAddr: a.config.GetArray("messaging.nsq.consumer_lookupd_addrs"),
		ProducerConfig:      pcfg,
		ConsumerConfig:      ccfg,
	})
	if err != nil {
		slog.Error("failed to init messaging", "error", err)
		os.Exit(1)
	}

	a.messaging = nsq
}

func (a *App) initHTTPServer() {
	a.router = pkgrouter.NewRouter(a.uuid, a.jwtAccessToken)

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
		"Messaging": func(context.Context) error {
			return a.messaging.Close()
		},
		"Config": func(context.Context) error {
			return a.config.Close()
		},
	}
}
