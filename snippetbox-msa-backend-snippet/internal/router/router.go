package router

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-snippet/internal/handler"
	"snippetbox/backend-snippet/internal/middleware"
)

const defaultMaxBodyBytes int64 = 1 << 20

type Config struct {
	Snippets       *handler.SnippetHandler
	System         *handler.SystemHandler
	Authenticator  middleware.Authenticator
	Logger         *slog.Logger
	AllowedOrigins []string
	MaxBodyBytes   int64
}

func New(cfg Config) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = defaultMaxBodyBytes
	}

	e := echo.New()
	e.HTTPErrorHandler = handler.NewErrorHandler(cfg.Logger)
	registerRoutes(e, cfg)

	return middleware.Chain(e,
		middleware.RequestID(),
		middleware.AccessLog(cfg.Logger),
		middleware.Recoverer(cfg.Logger),
		middleware.SecureHeaders(),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.BodyLimit(cfg.MaxBodyBytes),
	)
}

func registerRoutes(e *echo.Echo, cfg Config) {
	e.GET("/ping", cfg.System.Ping)
	e.GET("/swagger", cfg.System.SwaggerRedirect)
	e.GET("/swagger/", cfg.System.SwaggerUI)
	e.GET("/swagger/openapi.json", cfg.System.OpenAPI)

	requireAuth := middleware.Auth(cfg.Authenticator)

	v1 := e.Group(handler.APIPrefix)

	v1.GET("/snippets", cfg.Snippets.List)
	v1.GET("/snippets/:id", cfg.Snippets.Get)
	v1.POST("/snippets", cfg.Snippets.Create, requireAuth)
}
