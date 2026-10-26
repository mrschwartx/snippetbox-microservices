package config_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/config"
)

const validSecret = "0123456789abcdef0123456789abcdef"

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func withSecret(extra map[string]string) map[string]string {
	values := map[string]string{"JWT_SECRET": validSecret}
	for key, value := range extra {
		values[key] = value
	}
	return values
}

func TestLoad_Defaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(env(withSecret(nil)))
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, ":8082", cfg.HTTPAddr)
	assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
	assert.Equal(t, 30, cfg.DBConnectRetries)
	assert.Equal(t, 2*time.Second, cfg.DBConnectRetryDelay)
	assert.Equal(t, validSecret, cfg.JWTSecret)
	assert.Equal(t, []string{"http://localhost:3000"}, cfg.CORSAllowedOrigins)

	parsed, err := mysql.ParseDSN(cfg.DBDSN)
	require.NoError(t, err)
	assert.Equal(t, "snippetbox", parsed.User)
	assert.Equal(t, "localhost:3306", parsed.Addr)
	assert.Equal(t, "snippetbox", parsed.DBName)
	assert.True(t, parsed.ParseTime)
	assert.Equal(t, time.UTC, parsed.Loc)
	assert.Contains(t, cfg.DBDSN, "charset=utf8mb4")
}

func TestLoad_Overrides(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(env(withSecret(map[string]string{
		"APP_ENV":                "production",
		"HTTP_ADDR":              " :9090 ",
		"LOG_LEVEL":              "debug",
		"DB_CONNECT_RETRIES":     "3",
		"DB_CONNECT_RETRY_DELAY": "500ms",
		"CORS_ALLOWED_ORIGINS":   " https://a.test , ,https://b.test ",
	})))
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, ":9090", cfg.HTTPAddr)
	assert.Equal(t, slog.LevelDebug, cfg.LogLevel)
	assert.Equal(t, 3, cfg.DBConnectRetries)
	assert.Equal(t, 500*time.Millisecond, cfg.DBConnectRetryDelay)
	assert.Equal(t, []string{"https://a.test", "https://b.test"}, cfg.CORSAllowedOrigins)
}

func TestLoad_DatabaseDSN(t *testing.T) {
	t.Parallel()

	t.Run("built from parts with escaping", func(t *testing.T) {
		t.Parallel()
		cfg, err := config.Load(env(withSecret(map[string]string{
			"DB_HOST": "db", "DB_PORT": "3307", "DB_USER": "app", "DB_PASSWORD": "p@ss/w:rd ", "DB_NAME": "snip",
		})))
		require.NoError(t, err)

		parsed, err := mysql.ParseDSN(cfg.DBDSN)
		require.NoError(t, err)
		assert.Equal(t, "app", parsed.User)
		assert.Equal(t, "p@ss/w:rd ", parsed.Passwd)
		assert.Equal(t, "db:3307", parsed.Addr)
		assert.Equal(t, "snip", parsed.DBName)
	})

	t.Run("explicit DSN wins", func(t *testing.T) {
		t.Parallel()
		cfg, err := config.Load(env(withSecret(map[string]string{
			"DB_DSN": " user:pw@tcp(h:1)/x ", "DB_HOST": "ignored",
		})))
		require.NoError(t, err)
		assert.Equal(t, "user:pw@tcp(h:1)/x", cfg.DBDSN)
	})
}

func TestLoad_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		values  map[string]string
		message string
	}{
		{"missing secret", map[string]string{}, "JWT_SECRET"},
		{"short secret", map[string]string{"JWT_SECRET": "short"}, "JWT_SECRET"},
		{"whitespace only secret", map[string]string{"JWT_SECRET": strings.Repeat(" ", 40)}, "JWT_SECRET"},
		{"bad log level", withSecret(map[string]string{"LOG_LEVEL": "loud"}), "LOG_LEVEL"},
		{"unparsable retry delay", withSecret(map[string]string{"DB_CONNECT_RETRY_DELAY": "soon"}), "DB_CONNECT_RETRY_DELAY"},
		{"negative retry delay", withSecret(map[string]string{"DB_CONNECT_RETRY_DELAY": "-1s"}), "DB_CONNECT_RETRY_DELAY"},
		{"non numeric retries", withSecret(map[string]string{"DB_CONNECT_RETRIES": "many"}), "DB_CONNECT_RETRIES"},
		{"zero retries", withSecret(map[string]string{"DB_CONNECT_RETRIES": "0"}), "DB_CONNECT_RETRIES"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := config.Load(env(tt.values))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}
