package app

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nsqio/go-nsq"
	libOTP "github.com/pquerna/otp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
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
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func (a *App) initConfig() {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "/config/config.yaml"
		if os.Getenv("LOCAL") == "true" {
			path = "./config/config.yaml"
		}
	}

	cfg, err := config.NewViper(path)
	if err != nil {
		slog.Error("failed to init config", "error", err)
		os.Exit(1)
	}

	//nolint:errcheck,gosec // ignore error
	os.Setenv("TZ", cfg.GetString("app.tz"))

	a.config = cfg
}

func (a *App) initInstrument() {
	ins, err := instrument.New(context.Background(), &instrument.Config{
		Enabled:          true,
		ServiceName:      a.config.GetString("instrument.service_name"),
		ServiceVersion:   a.config.GetString("instrument.service_version"),
		Environment:      a.config.GetString("instrument.env"),
		OTLPEndpoint:     a.config.GetString("instrument.otlp_endpoint"),
		OTLPSecure:       a.config.GetBool("instrument.otlp_secure"),
		TraceSampleRatio: a.config.GetFloat64("instrument.trace_sample_ratio"),
		MetricsInterval:  a.config.GetSecond("instrument.metric_interval_seconds"),
		MaskFields:       a.config.GetArray("instrument.log_mask_fields"),
	})
	if err != nil {
		slog.Error("failed to init instrumentation", "error", err)
		os.Exit(1)
	}
	a.ins = ins
}

func (a *App) initLibraries() {
	a.clock = clock.New()
	a.uuid = uid.NewUUID()
	a.goroutine = goroutine.NewManager(a.config.GetInt("app.server.max_goroutine"))
	a.hmac = hash.NewHMACSHA256(a.config.GetString("hash.hmac.secret"))
	a.argon2id = hash.NewArgon2id(a.config.GetString("hash.argon2id.pepper"))
	a.bcrypt = hash.NewBcrypt(a.config.GetInt("hash.bcrypt.cost"), a.config.GetString("hash.bcrypt.pepper"))

	validator, err := validator.NewV10Validator()
	if err != nil {
		slog.Error("failed to init validation v10 validator", "error", err)
		os.Exit(1)
	}
	a.validator = validator

	snow, err := uid.NewSnowflake()
	if err != nil {
		slog.Error("failed to init uid number snowflake", "error", err)
		os.Exit(1)
	}
	a.uid = snow

	objID, err := uid.NewObjectIDGenerator()
	if err != nil {
		slog.Error("failed to init uid string object_id", "error", err)
		os.Exit(1)
	}
	a.oid = objID

	a.totp = otp.NewTOTP(
		a.config.GetString("mfa.totp.issuer"),
		a.config.GetUint("mfa.totp.period"),
		a.config.GetUint("mfa.totp.skew"),
		libOTP.DigitsSix,
	)

	rawKey, err := base64.StdEncoding.DecodeString(a.config.GetString("mfa.secret"))
	if err != nil {
		slog.Error("failed to decode mfa secret", "error", err)
		os.Exit(1)
	}
	if len(rawKey) != 32 {
		slog.Error("failed to init mfacrypto, secret must be 32 bytes (AES-256)", "error", err)
		os.Exit(1)
	}
	a.mfaEncryptor = mfa.NewAESGCMEncryptor(mfa.StaticKeyProvider{KeyBytes: rawKey})
	a.mfaRecoveryCode = mfa.NewRecoveryCode()
}

func (a *App) initJWT() {
	defaultJWT, err := jwt.NewHS512(jwt.Config{
		Secret:     []byte(a.config.GetString("jwt.secret")),
		Issuer:     a.config.GetString("jwt.issuer"),
		Audiences:  a.config.GetArray("jwt.audiences"),
		TTLMinutes: a.config.GetMinute("jwt.ttl_minutes"),
		Clock:      a.clock,
		UUID:       a.uuid,
	})
	if err != nil {
		slog.Error("failed to init jwt token", "error", err)
		os.Exit(1)
	}
	a.jwt = defaultJWT
}

func (a *App) initDatabase() {
	config, err := pgxpool.ParseConfig(a.config.GetString("database.url"))
	if err != nil {
		slog.Error("failed to parse DB connection string.", "error", err)
		os.Exit(1)
	}

	config.MaxConns = a.config.GetInt32("database.pool.max_conns")
	config.MinConns = a.config.GetInt32("database.pool.min_conns")
	config.MaxConnLifetime = a.config.GetSecond("database.pool.max_conn_lifetime_seconds")
	config.MaxConnIdleTime = a.config.GetSecond("database.pool.max_conn_idle_seconds")
	config.HealthCheckPeriod = a.config.GetSecond("database.pool.health_check_period_seconds")

	pool, err := pgxpool.NewWithConfig(a.ctx, config)
	if err != nil {
		slog.Error("failed to create DB connection pool", "error", err)
		os.Exit(1)
	}

	pingCtx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		slog.Error("failed to ping DB", "error", err)
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

	pingCtx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		slog.Error("failed to init redis", "error", err)
		os.Exit(1)
	}

	a.cacheConn = rdb
	a.idemp = idempotency.New(a.cacheConn)
}

func (a *App) initMail() {
	mail, err := mail.NewSMTP(mail.SMTPConfig{
		Host:     a.config.GetString("mail.host"),
		Port:     a.config.GetInt("mail.port"),
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

//nolint:gocognit // it's fine
func (a *App) initStorage() {
	driver := strings.TrimSpace(a.config.GetString("storage.driver"))

	var gcsClient *gcs.Client
	if driver == storage.DriverGCS {
		gcsOptions := []option.ClientOption{}
		if a.config.GetBool("storage.gcs.without_auth") {
			gcsOptions = append(gcsOptions, option.WithoutAuthentication())
		}
		if v := strings.TrimSpace(a.config.GetString("storage.gcs.credentials_file")); v != "" {
			// #nosec G304 -- path is from trusted config file.
			credsJSON, err := os.ReadFile(v)
			if err != nil {
				slog.Error("failed to read gcs credentials file", "error", err)
				os.Exit(1)
			}
			creds, err := google.CredentialsFromJSON(a.ctx, credsJSON, gcs.ScopeFullControl)
			if err != nil {
				slog.Error("failed to parse gcs credentials file", "error", err)
				os.Exit(1)
			}
			gcsOptions = append(gcsOptions, option.WithCredentials(creds))
		}
		if v := a.config.GetBinary("storage.gcs.credentials_json"); len(v) > 0 {
			creds, err := google.CredentialsFromJSON(a.ctx, v, gcs.ScopeFullControl)
			if err != nil {
				slog.Error("failed to parse gcs credentials json", "error", err)
				os.Exit(1)
			}
			gcsOptions = append(gcsOptions, option.WithCredentials(creds))
		}
		if v := strings.TrimSpace(a.config.GetString("storage.gcs.endpoint")); v != "" {
			gcsOptions = append(gcsOptions, option.WithEndpoint(v))
		}
		if v := strings.TrimSpace(a.config.GetString("storage.gcs.user_agent")); v != "" {
			gcsOptions = append(gcsOptions, option.WithUserAgent(v))
		}
		if len(gcsOptions) > 0 {
			client, err := gcs.NewClient(a.ctx, gcsOptions...)
			if err != nil {
				slog.Error("failed to init gcs client", "error", err)
				os.Exit(1)
			}
			gcsClient = client
		}
	}

	stg, err := storage.NewFromDriver(a.ctx, driver, storage.FactoryOptions{
		S3: storage.S3Options{
			Region:       strings.TrimSpace(a.config.GetString("storage.s3.region")),
			Endpoint:     strings.TrimSpace(a.config.GetString("storage.s3.endpoint")),
			AccessKey:    strings.TrimSpace(a.config.GetString("storage.s3.access_key")),
			SecretKey:    strings.TrimSpace(a.config.GetString("storage.s3.secret_key")),
			SessionToken: strings.TrimSpace(a.config.GetString("storage.s3.session_token")),
			UsePathStyle: a.config.GetBool("storage.s3.use_path_style"),
		},
		GCS: storage.GCSOptions{
			Client:         gcsClient,
			GoogleAccessID: strings.TrimSpace(a.config.GetString("storage.gcs.signer_access_id")),
			PrivateKey:     a.config.GetBinary("storage.gcs.signer_private_key"),
		},
		MinIO: storage.MinIOOptions{
			Region:       strings.TrimSpace(a.config.GetString("storage.minio.region")),
			Endpoint:     strings.TrimSpace(a.config.GetString("storage.minio.endpoint")),
			AccessKey:    strings.TrimSpace(a.config.GetString("storage.minio.access_key")),
			SecretKey:    strings.TrimSpace(a.config.GetString("storage.minio.secret_key")),
			SessionToken: strings.TrimSpace(a.config.GetString("storage.minio.session_token")),
			UseSSL:       a.config.GetBool("storage.minio.use_ssl"),
		},
	})
	if err != nil {
		slog.Error("failed to init storage", "error", err)
		os.Exit(1)
	}

	a.storage = stg
}

func (a *App) initMessaging() {
	driver := a.config.GetString("messaging.driver")
	client, err := messaging.NewFromDriver(a.ctx, driver, messaging.FactoryOptions{
		NSQ: messaging.NSQConfig{
			ProducerAddr:         a.config.GetString("messaging.nsq.producer_addr"),
			ConsumerNSQDAddrs:    a.config.GetArray("messaging.nsq.consumer_nsqd_addrs"),
			ConsumerLookupdAddrs: a.config.GetArray("messaging.nsq.consumer_lookupd_addrs"),
			ProducerConfig: func() *nsq.Config {
				cfg := nsq.NewConfig()
				cfg.MaxInFlight = a.config.GetInt("messaging.nsq.producer_config.max_in_flight")
				cfg.DialTimeout = a.config.GetSecond("messaging.nsq.producer_config.dial_timeout_seconds")
				cfg.ReadTimeout = a.config.GetSecond("messaging.nsq.producer_config.read_timeout_seconds")
				cfg.WriteTimeout = a.config.GetSecond("messaging.nsq.producer_config.write_timeout_seconds")
				return cfg
			}(),
			ConsumerConfig: func() *nsq.Config {
				cfg := nsq.NewConfig()
				cfg.MaxInFlight = a.config.GetInt("messaging.nsq.consumer_config.max_in_flight")
				cfg.MaxAttempts = a.config.GetUint16("messaging.nsq.consumer_config.max_attempts")
				cfg.LookupdPollInterval = a.config.GetSecond("messaging.nsq.consumer_config.lookupd_poll_interval_seconds")
				cfg.DialTimeout = a.config.GetSecond("messaging.nsq.consumer_config.dial_timeout_seconds")
				cfg.ReadTimeout = a.config.GetSecond("messaging.nsq.consumer_config.read_timeout_seconds")
				cfg.WriteTimeout = a.config.GetSecond("messaging.nsq.consumer_config.write_timeout_seconds")
				cfg.DefaultRequeueDelay = a.config.GetSecond("messaging.nsq.consumer_config.default_requeue_delay_seconds")
				cfg.MaxRequeueDelay = a.config.GetSecond("messaging.nsq.consumer_config.max_requeue_delay_seconds")
				return cfg
			}(),
		},
		NATS: messaging.NATSConfig{
			URL: a.config.GetString("messaging.nats.url"),
			Options: []nats.Option{
				nats.Name(a.config.GetString("messaging.nats.name")),
				nats.MaxReconnects(a.config.GetInt("messaging.nats.max_reconnects")),
				nats.Timeout(a.config.GetSecond("messaging.nats.timeout_seconds")),
				nats.ReconnectWait(a.config.GetSecond("messaging.nats.reconnect_wait_seconds")),
				nats.PingInterval(a.config.GetSecond("messaging.nats.ping_interval_seconds")),
				nats.MaxPingsOutstanding(a.config.GetInt("messaging.nats.max_pings_outstanding")),
				nats.RetryOnFailedConnect(a.config.GetBool("messaging.nats.retry_on_failed_connect")),
				// nats.NoEcho(), if a.config.GetBool("messaging.nats.no_echo") == true
			},
		},
	})
	if err != nil {
		slog.Error("failed to init messaging", "error", err, "driver", driver)
		os.Exit(1)
	}

	a.messaging = client
}

func (a *App) initCasbin() {
	const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (p.obj == "*" || r.obj == p.obj) && (p.act == "*" || r.act == p.act)
`
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		slog.Error("failed to create model casbin", "error", err)
		os.Exit(1)
	}

	adapter, err := pgxcasbin.NewAdapter(a.ctx, a.dbConn, pgxcasbin.WithTableName("identity_casbin_rules"))
	if err != nil {
		slog.Error("failed to create adapter casbin", "error", err)
		os.Exit(1)
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		slog.Error("failed to init casbin", "error", err)
		os.Exit(1)
	}

	watcher, err := pgxcasbin.NewWatcherWithPool(a.ctx, a.dbConn,
		pgxcasbin.OptionWatcher{
			NotifySelf: true,
			Channel:    "iam_casbin_psql_watcher",
			Verbose:    false,
			LocalID:    a.uuid.Generate(),
		},
	)
	if err != nil {
		slog.Error("failed to create watcher casbin", "error", err)
		os.Exit(1)
	}

	if err := watcher.SetUpdateCallback(pgxcasbin.DefaultCallback(e)); err != nil {
		slog.Error("failed to create watcher fallback casbin", "error", err)
		os.Exit(1)
	}

	if err := e.SetWatcher(watcher); err != nil {
		slog.Error("failed to set watcher casbin", "error", err)
		os.Exit(1)
	}

	e.EnableAutoSave(true)
	e.EnableAutoNotifyWatcher(true)

	a.casbin = e
	a.casbinWatcher = watcher
}

func (a *App) initHTTPServer() {
	a.router = router.NewRouter(router.Config{
		Config:     a.config,
		UUID:       a.uuid,
		JWT:        a.jwt,
		Instrument: a.ins,
		Enforcer:   a.casbin,
	})

	routerWithCORS := cors.New(cors.Options{
		AllowedOrigins: a.config.GetArray("app.server.cors"),
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(a.router)

	a.httpServer = &http.Server{
		Addr:              a.config.GetString("app.server.http.address"),
		Handler:           routerWithCORS,
		ReadTimeout:       a.config.GetSecond("app.server.http.read_timeout_seconds"),
		ReadHeaderTimeout: a.config.GetSecond("app.server.http.read_header_timeout_seconds"),
		WriteTimeout:      a.config.GetSecond("app.server.http.write_timeout_seconds"),
		IdleTimeout:       a.config.GetSecond("app.server.http.idle_timeout_seconds"),
	}

	a.sseServer = &http.Server{
		Addr:              a.config.GetString("app.server.sse.address"),
		Handler:           routerWithCORS,
		ReadHeaderTimeout: a.config.GetSecond("app.server.sse.read_header_timeout_seconds"),
	}
}

func (a *App) initClosers() {
	a.closers = []struct {
		name string
		fn   func(context.Context) error
	}{
		{
			name: "Instrument",
			fn: func(ctx context.Context) error {
				return a.ins.Shutdown(ctx)
			},
		},
		{
			name: "Messaging",
			fn: func(context.Context) error {
				return a.messaging.Close()
			},
		},
		{
			name: "CasbinWatcher",
			fn: func(context.Context) error {
				if a.casbinWatcher != nil {
					a.casbinWatcher.Close()
				}

				return nil
			},
		},
		{
			name: "Redis",
			fn: func(context.Context) error {
				return a.cacheConn.Close()
			},
		},
		{
			name: "Database",
			fn: func(context.Context) error {
				a.dbConn.Close()

				return nil
			},
		},
		{
			name: "Storage",
			fn: func(context.Context) error {
				return a.storage.Close()
			},
		},
		{
			name: "Config",
			fn: func(context.Context) error {
				return a.config.Close()
			},
		},
	}
}
