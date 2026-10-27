package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const minJWTSecretLength = 32

type Config struct {
	Env                 string
	HTTPAddr            string
	LogLevel            slog.Level
	DBDSN               string
	DBConnectRetries    int
	DBConnectRetryDelay time.Duration
	JWTSecret           string
	CORSAllowedOrigins  []string
}

type lookup func(key string) string

func (l lookup) get(key, fallback string) string {
	if value := strings.TrimSpace(l(key)); value != "" {
		return value
	}
	return fallback
}

func Load(getenv func(string) string) (Config, error) {
	env := lookup(getenv)

	cfg := Config{
		Env:      env.get("APP_ENV", "development"),
		HTTPAddr: env.get("HTTP_ADDR", ":8082"),
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(env.get("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("config: invalid LOG_LEVEL: %w", err)
	}
	if err := cfg.loadAuth(env); err != nil {
		return Config{}, err
	}
	if err := cfg.loadDatabase(env); err != nil {
		return Config{}, err
	}
	cfg.CORSAllowedOrigins = splitOrigins(env.get("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))

	return cfg, nil
}

func (c *Config) loadAuth(env lookup) error {
	c.JWTSecret = strings.TrimSpace(env("JWT_SECRET"))
	if len(c.JWTSecret) < minJWTSecretLength {
		return fmt.Errorf("config: JWT_SECRET is required and must be at least %d characters", minJWTSecretLength)
	}

	return nil
}

func (c *Config) loadDatabase(env lookup) error {
	c.DBDSN = strings.TrimSpace(env("DB_DSN"))
	if c.DBDSN == "" {
		c.DBDSN = buildDSN(
			env.get("DB_HOST", "localhost"),
			env.get("DB_PORT", "3306"),
			env.get("DB_USER", "snippetbox"),
			env("DB_PASSWORD"),
			env.get("DB_NAME", "snippetbox"),
		)
	}

	delay, err := time.ParseDuration(env.get("DB_CONNECT_RETRY_DELAY", "2s"))
	if err != nil || delay < 0 {
		return errors.New("config: DB_CONNECT_RETRY_DELAY must be a duration such as 2s")
	}
	c.DBConnectRetryDelay = delay

	retries, err := strconv.Atoi(env.get("DB_CONNECT_RETRIES", "30"))
	if err != nil || retries < 1 {
		return errors.New("config: DB_CONNECT_RETRIES must be a positive integer")
	}
	c.DBConnectRetries = retries

	return nil
}

func splitOrigins(raw string) []string {
	var origins []string
	for _, origin := range strings.Split(raw, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func buildDSN(host, port, user, password, name string) string {
	c := mysql.NewConfig()
	c.User = user
	c.Passwd = password
	c.Net = "tcp"
	c.Addr = net.JoinHostPort(host, port)
	c.DBName = name
	c.ParseTime = true
	c.Loc = time.UTC
	c.Params = map[string]string{"charset": "utf8mb4"}
	return c.FormatDSN()
}
