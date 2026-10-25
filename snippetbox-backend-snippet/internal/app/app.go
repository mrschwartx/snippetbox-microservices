package app

import (
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"snippetbox/backend-snippet/docs"
	"snippetbox/backend-snippet/internal/config"
	"snippetbox/backend-snippet/internal/handler"
	"snippetbox/backend-snippet/internal/repository"
	"snippetbox/backend-snippet/internal/router"
	"snippetbox/backend-snippet/internal/security"
	"snippetbox/backend-snippet/internal/service"
)

type Dependencies struct {
	Config config.Config
	DB     *gorm.DB
	Logger *slog.Logger
	Clock  service.Clock
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

func NewHandler(d Dependencies) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	if d.Clock == nil {
		d.Clock = systemClock{}
	}

	authenticator := service.NewTokenAuthenticator(security.NewJWTVerifier(d.Config.JWTSecret), d.Clock)
	snippetService := service.NewSnippetService(repository.NewSnippetRepository(d.DB), d.Clock)

	return router.New(router.Config{
		Snippets:       handler.NewSnippetHandler(snippetService),
		System:         handler.NewSystemHandler(docs.OpenAPI, docs.SwaggerHTML),
		Authenticator:  authenticator,
		Logger:         d.Logger,
		AllowedOrigins: d.Config.CORSAllowedOrigins,
	})
}
