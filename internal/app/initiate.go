package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nsqio/go-nsq"
	"github.com/pquerna/otp"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
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
	"google.golang.org/api/option"
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

//nolint:gocognit,gocyclo,cyclop,errcheck // it's fine
func (a *App) initMessaging() {
	driver := strings.ToLower(strings.TrimSpace(a.config.GetString("messaging.driver")))
	if driver == "" {
		driver = "nsq"
	}

	var (
		client pkgmessaging.Messaging
		err    error
	)

	switch driver {
	case "nsq":
		// consumer config
		ccfg := nsq.NewConfig()
		if v := a.config.GetInt("messaging.nsq.consumer_config.max_in_flight"); v > 0 {
			ccfg.MaxInFlight = int(v)
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.lookupd_poll_interval"); v > 0 {
			ccfg.LookupdPollInterval = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.dial_timeout"); v > 0 {
			ccfg.DialTimeout = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.read_timeout"); v > 0 {
			ccfg.ReadTimeout = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.write_timeout"); v > 0 {
			ccfg.WriteTimeout = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.max_attempts"); v > 0 && v <= 65535 {
			ccfg.MaxAttempts = uint16(v)
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.msg_timeout"); v > 0 {
			ccfg.MsgTimeout = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.default_requeue_delay"); v > 0 {
			ccfg.DefaultRequeueDelay = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.consumer_config.max_requeue_delay"); v > 0 {
			ccfg.MaxRequeueDelay = time.Duration(v) * time.Second
		}

		// producer config
		pcfg := nsq.NewConfig()
		if v := a.config.GetInt("messaging.nsq.producer_config.max_in_flight"); v > 0 {
			pcfg.MaxInFlight = int(v)
		}
		if v := a.config.GetInt("messaging.nsq.producer_config.dial_timeout"); v > 0 {
			pcfg.DialTimeout = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.producer_config.read_timeout"); v > 0 {
			pcfg.ReadTimeout = time.Duration(v) * time.Second
		}
		if v := a.config.GetInt("messaging.nsq.producer_config.write_timeout"); v > 0 {
			pcfg.WriteTimeout = time.Duration(v) * time.Second
		}

		cLookupdAddrs := a.config.GetArray("messaging.nsq.consumer_lookupd_addrs")
		cNSQDAddrs := a.config.GetArray("messaging.nsq.consumer_nsqd_addrs")
		if len(cLookupdAddrs) == 0 && len(cNSQDAddrs) == 0 {
			slog.Error("failed to init messaging: empty cLookupdAddrs and cNSQDAddrs")
			os.Exit(1)
		}

		client, err = pkgmessaging.NewNSQ(pkgmessaging.NSQConfig{
			ProducerAddr:         a.config.GetString("messaging.nsq.producer_addr"),
			ConsumerNSQDAddrs:    cNSQDAddrs,
			ConsumerLookupdAddrs: cLookupdAddrs,
			ProducerConfig:       pcfg,
			ConsumerConfig:       ccfg,
		})
	case "kafka":
		dialerTimeout := a.config.GetInt("messaging.kafka.dial_timeout")
		var dialer *kafka.Dialer
		if dialerTimeout > 0 {
			dialer = &kafka.Dialer{Timeout: time.Duration(dialerTimeout) * time.Second}
		}

		var writerCfg *kafka.WriterConfig
		{
			var wc kafka.WriterConfig
			set := false

			if v := a.config.GetInt("messaging.kafka.writer_config.max_attempts"); v > 0 {
				wc.MaxAttempts = int(v)
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.writer_config.batch_size"); v > 0 {
				wc.BatchSize = int(v)
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.writer_config.batch_bytes"); v > 0 {
				wc.BatchBytes = int(v)
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.writer_config.batch_timeout"); v > 0 {
				wc.BatchTimeout = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.writer_config.read_timeout"); v > 0 {
				wc.ReadTimeout = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.writer_config.write_timeout"); v > 0 {
				wc.WriteTimeout = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.writer_config.required_acks"); v != 0 {
				wc.RequiredAcks = int(v)
				set = true
			}
			if a.config.GetBool("messaging.kafka.writer_config.async") {
				wc.Async = true
				set = true
			}

			if set {
				writerCfg = &wc
			}
		}

		var readerCfg *kafka.ReaderConfig
		{
			var rc kafka.ReaderConfig
			set := false

			if v := a.config.GetInt("messaging.kafka.reader_config.queue_capacity"); v > 0 {
				rc.QueueCapacity = int(v)
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.min_bytes"); v > 0 {
				rc.MinBytes = int(v)
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.max_bytes"); v > 0 {
				rc.MaxBytes = int(v)
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.max_wait"); v > 0 {
				rc.MaxWait = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.read_batch_timeout"); v > 0 {
				rc.ReadBatchTimeout = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.max_attempts"); v > 0 {
				rc.MaxAttempts = int(v)
				set = true
			}
			if a.config.GetBool("messaging.kafka.reader_config.watch_partition_changes") {
				rc.WatchPartitionChanges = true
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.partition_watch_interval"); v > 0 {
				rc.PartitionWatchInterval = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.session_timeout"); v > 0 {
				rc.SessionTimeout = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.rebalance_timeout"); v > 0 {
				rc.RebalanceTimeout = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.heartbeat_interval"); v > 0 {
				rc.HeartbeatInterval = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.join_group_backoff"); v > 0 {
				rc.JoinGroupBackoff = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.retention_time"); v != 0 {
				rc.RetentionTime = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.start_offset"); v != 0 {
				rc.StartOffset = v
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.read_backoff_min"); v > 0 {
				rc.ReadBackoffMin = time.Duration(v) * time.Second
				set = true
			}
			if v := a.config.GetInt("messaging.kafka.reader_config.read_backoff_max"); v > 0 {
				rc.ReadBackoffMax = time.Duration(v) * time.Second
				set = true
			}

			if set {
				readerCfg = &rc
			}
		}

		kafkaBrokers := a.config.GetArray("messaging.kafka.brokers")
		if len(kafkaBrokers) == 0 {
			slog.Error("failed to init messaging: empty kafkaBrokers")
			os.Exit(1)
		}

		client, err = pkgmessaging.NewKafka(pkgmessaging.KafkaConfig{
			Brokers: kafkaBrokers,
			Dialer:  dialer,

			WriterConfig: writerCfg,
			ReaderConfig: readerCfg,
		})
	case "nats":
		var opts []nats.Option
		if v := strings.TrimSpace(a.config.GetString("messaging.nats.name")); v != "" {
			opts = append(opts, nats.Name(v))
		}
		if user := strings.TrimSpace(a.config.GetString("messaging.nats.user")); user != "" {
			if pass := a.config.GetString("messaging.nats.password"); pass != "" {
				opts = append(opts, nats.UserInfo(user, pass))
			}
		}
		if token := strings.TrimSpace(a.config.GetString("messaging.nats.token")); token != "" {
			opts = append(opts, nats.Token(token))
		}
		if v := a.config.GetInt("messaging.nats.timeout"); v > 0 {
			opts = append(opts, nats.Timeout(time.Duration(v)*time.Second))
		}
		if v := a.config.GetInt("messaging.nats.max_reconnects"); v != 0 {
			opts = append(opts, nats.MaxReconnects(int(v)))
		}
		if v := a.config.GetInt("messaging.nats.reconnect_wait"); v > 0 {
			opts = append(opts, nats.ReconnectWait(time.Duration(v)*time.Second))
		}
		jitter := a.config.GetInt("messaging.nats.reconnect_jitter")
		jitterTLS := a.config.GetInt("messaging.nats.reconnect_jitter_tls")
		if jitter > 0 || jitterTLS > 0 {
			opts = append(opts, nats.ReconnectJitter(time.Duration(jitter)*time.Second, time.Duration(jitterTLS)*time.Second))
		}
		if v := a.config.GetInt("messaging.nats.ping_interval"); v > 0 {
			opts = append(opts, nats.PingInterval(time.Duration(v)*time.Second))
		}
		if v := a.config.GetInt("messaging.nats.max_pings_outstanding"); v > 0 {
			opts = append(opts, nats.MaxPingsOutstanding(int(v)))
		}
		if a.config.GetBool("messaging.nats.no_echo") {
			opts = append(opts, nats.NoEcho())
		}
		if v := strings.TrimSpace(a.config.GetString("messaging.nats.inbox_prefix")); v != "" {
			opts = append(opts, nats.CustomInboxPrefix(v))
		}
		opts = append(opts, nats.RetryOnFailedConnect(a.config.GetBool("messaging.nats.retry_on_failed_connect")))

		client, err = pkgmessaging.NewNATS(pkgmessaging.NATSConfig{
			URL:     a.config.GetString("messaging.nats.url"),
			Options: opts,
		})
	case "pubsub":
		var opts []option.ClientOption
		if a.config.GetBool("messaging.pubsub.without_auth") {
			opts = append(opts, option.WithoutAuthentication())
		}
		if v := strings.TrimSpace(a.config.GetString("messaging.pubsub.credentials_file")); v != "" {
			opts = append(opts, option.WithCredentialsFile(v))
		}
		if v := a.config.GetBinary("messaging.pubsub.credentials_json"); len(v) > 0 {
			opts = append(opts, option.WithCredentialsJSON(v))
		}
		if v := strings.TrimSpace(a.config.GetString("messaging.pubsub.endpoint")); v != "" {
			opts = append(opts, option.WithEndpoint(v))
		}
		if v := strings.TrimSpace(a.config.GetString("messaging.pubsub.user_agent")); v != "" {
			opts = append(opts, option.WithUserAgent(v))
		}
		if v := strings.TrimSpace(a.config.GetString("messaging.pubsub.emulator_host")); v != "" {
			_ = os.Setenv("PUBSUB_EMULATOR_HOST", v)
		}

		client, err = pkgmessaging.NewPubSub(a.ctx, pkgmessaging.PubSubConfig{
			ProjectID:     a.config.GetString("messaging.pubsub.project_id"),
			ClientOptions: opts,
		})
	default:
		slog.Error("failed to init messaging: unknown driver", "driver", driver)
		os.Exit(1)
	}

	if err != nil {
		slog.Error("failed to init messaging", "error", err, "driver", driver)
		os.Exit(1)
	}

	a.messaging = client
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
