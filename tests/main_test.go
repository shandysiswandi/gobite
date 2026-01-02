package tests

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/app"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testEnv struct {
	baseURL string
	pool    *pgxpool.Pool
}

var (
	env     testEnv
	cleanup func()
)

// BE CAREFUL: this TestMain is use sharing the same env and DB across tests.
// Parallel tests are safe only if they don’t mutate shared state or you isolate data per test
func TestMain(m *testing.M) {
	_ = os.Setenv("LOCAL", "true")

	var err error
	env, cleanup, err = setupIntegration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup integration: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	if cleanup != nil {
		cleanup()
	}
	os.Exit(code)
}

func setupIntegration() (testEnv, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	netw, err := network.New(ctx)
	if err != nil {
		cancel()
		return testEnv{}, nil, fmt.Errorf("create network: %w", err)
	}
	networkName := netw.Name

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("gobite"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
	)
	if err != nil {
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("start postgres: %w", err)
	}

	redisContainer, err := redis.Run(ctx, "redis:8.2-alpine")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("start redis: %w", err)
	}

	mailerReq := testcontainers.ContainerRequest{
		Image:        "axllent/mailpit:v1.28.0",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForListeningPort("1025/tcp"),
	}
	mailerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: mailerReq,
		Started:          true,
	})
	if err != nil {
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("start mailer: %w", err)
	}

	lookupdReq := testcontainers.ContainerRequest{
		Image:        "nsqio/nsq:v1.3.0",
		ExposedPorts: []string{"4160/tcp", "4161/tcp"},
		Cmd:          []string{"/nsqlookupd"},
		WaitingFor:   wait.ForListeningPort("4160/tcp"),
		Networks:     []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"nsqlookupd"},
		},
	}
	lookupdContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: lookupdReq,
		Started:          true,
	})
	if err != nil {
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("start nsqlookupd: %w", err)
	}

	nsqdReq := testcontainers.ContainerRequest{
		Image:        "nsqio/nsq:v1.3.0",
		ExposedPorts: []string{"4150/tcp", "4151/tcp"},
		Cmd: []string{
			"/nsqd",
			"--lookupd-tcp-address=nsqlookupd:4160",
			"--broadcast-address=localhost",
			"--data-path=/data",
		},
		WaitingFor: wait.ForListeningPort("4150/tcp"),
		Networks:   []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"nsqd"},
		},
	}
	nsqdContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: nsqdReq,
		Started:          true,
	})
	if err != nil {
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("start nsqd: %w", err)
	}

	pgURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("postgres connection string: %w", err)
	}

	redisEndpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("redis endpoint: %w", err)
	}
	redisURL := fmt.Sprintf("redis://%s/1", redisEndpoint)

	mailerAddr, err := mappedAddr(ctx, mailerContainer, "1025/tcp")
	if err != nil {
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("mailer addr: %w", err)
	}

	lookupdAddr, err := mappedAddr(ctx, lookupdContainer, "4161/tcp")
	if err != nil {
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("lookupd addr: %w", err)
	}

	nsqdAddr, err := mappedAddr(ctx, nsqdContainer, "4150/tcp")
	if err != nil {
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("nsqd addr: %w", err)
	}

	pool, err := pgxpool.New(ctx, pgURL)
	if err != nil {
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("connect postgres: %w", err)
	}

	migrationsDir, err := migrationsPath()
	if err != nil {
		pool.Close()
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("migrations path: %w", err)
	}
	if err := applyMigrations(ctx, pool, migrationsDir); err != nil {
		pool.Close()
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("apply migrations: %w", err)
	}

	configPath, err := writeTestConfig(pgURL, redisURL, mailerAddr, lookupdAddr, nsqdAddr)
	if err != nil {
		pool.Close()
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("write config: %w", err)
	}
	if err := os.Setenv("CONFIG_PATH", configPath); err != nil {
		pool.Close()
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("set CONFIG_PATH: %w", err)
	}

	application := app.New()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		pool.Close()
		_ = nsqdContainer.Terminate(ctx)
		_ = lookupdContainer.Terminate(ctx)
		_ = mailerContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
		_ = pgContainer.Terminate(ctx)
		_ = netw.Remove(ctx)
		cancel()
		return testEnv{}, nil, fmt.Errorf("listen: %w", err)
	}
	errChan := application.Serve(listener)
	baseURL := "http://" + listener.Addr().String()

	cleanup := func() {
		pool.Close()

		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		application.Stop(stopCtx)

		select {
		case err := <-errChan:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			}
		case <-time.After(5 * time.Second):
			fmt.Fprintln(os.Stderr, "server shutdown timed out")
		}

		_ = nsqdContainer.Terminate(stopCtx)
		_ = lookupdContainer.Terminate(stopCtx)
		_ = mailerContainer.Terminate(stopCtx)
		_ = redisContainer.Terminate(stopCtx)
		_ = pgContainer.Terminate(stopCtx)
		_ = netw.Remove(stopCtx)
		cancel()
	}

	return testEnv{
		baseURL: baseURL,
		pool:    pool,
	}, cleanup, nil
}

func writeTestConfig(pgURL, redisURL, mailAddr, lookupdAddr, nsqdAddr string) (string, error) {
	dir, err := os.MkdirTemp("", "gobite-config-*")
	if err != nil {
		return "", err
	}

	mailHost, mailPort, err := net.SplitHostPort(mailAddr)
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(dir, "config.yaml")
	configBody := fmt.Sprintf(`
app:
  tz: "UTC"
  web: "http://localhost:5173"
  server:
    http:
      address: "127.0.0.1:0"

instrument:
  log_mask_fields: "password,new_password,current_password,access_token,refresh_token,authorization,cookie,recovery_codes"

database:
  url: "%s"

redis:
  url: "%s"

mail:
  host: "%s"
  port: %s
  username: "user"
  password: "password"
  from: "no-reply@gobite.com"

messaging:
  driver: nsq
  nsq:
    producer_addr: "%s"
    consumer_nsqd_addrs: ""
    consumer_lookupd_addrs: "%s"

jwt:
  secret: "test-secret-test-secret-test-secret-test-secret-test-secret-test-secret-64b"
  issuer: "gobite-test"
  audiences: "WEB"
  ttl: 5

hash:
  hmac:
    secret: "test-hmac-secret"
  argon2id:
    pepper: "pepper"
  bcrypt:
    cost: 4
    pepper: "pepper"

mfa:
  secret: "zWekm8jt0dAiUhMbQ3Sz5vx4KF9bv9iaPYVwpjlSafU="
  totp:
    issuer: "GOBITE"
    period: 30
    skew: 1

modules:
  identity:
    enabled: true
    registration_ttl: 3
    password_reset_ttl: 3
    refresh_token_ttl: 7
    mfa_login_ttl: 5
    mfa_setup_confirm_ttl: 5
  iam:
    enabled: true
  notification:
    enabled: false
    consumer_names: ""
`, pgURL, redisURL, mailHost, mailPort, nsqdAddr, lookupdAddr)

	if err := os.WriteFile(configPath, []byte(configBody), 0o600); err != nil {
		return "", err
	}

	return configPath, nil
}

func migrationsPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root := filepath.Dir(wd)
	return filepath.Join(root, "database", "migrations"), nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)

	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		sql := extractUpSQL(string(content))
		if strings.TrimSpace(sql) == "" {
			continue
		}
		if _, err := pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("migrate %s: %w", filepath.Base(filePath), err)
		}
	}

	return nil
}

func mappedAddr(ctx context.Context, c testcontainers.Container, portID string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	natPort, err := parseNatPort(portID)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, natPort)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(host, port.Port()), nil
}

func parseNatPort(portID string) (nat.Port, error) {
	parts := strings.Split(portID, "/")
	if len(parts) == 2 {
		return nat.NewPort(parts[1], parts[0])
	}

	return nat.NewPort("tcp", portID)
}

func extractUpSQL(content string) string {
	hasStatements := strings.Contains(content, "-- +goose StatementBegin")
	var sb strings.Builder

	inUp := false
	inStatement := false
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		switch strings.TrimSpace(line) {
		case "-- +goose Up":
			inUp = true
			inStatement = false
			continue
		case "-- +goose Down":
			inUp = false
			inStatement = false
			continue
		case "-- +goose StatementBegin":
			if inUp {
				inStatement = true
			}
			continue
		case "-- +goose StatementEnd":
			if inUp {
				inStatement = false
			}
			continue
		}

		if inUp && (inStatement || !hasStatements) {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return strings.TrimSpace(sb.String())
}
