package app

import (
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"snippetbox/backend-user/docs"
	"snippetbox/backend-user/internal/config"
	"snippetbox/backend-user/internal/handler"
	"snippetbox/backend-user/internal/repository"
	"snippetbox/backend-user/internal/router"
	"snippetbox/backend-user/internal/security"
	"snippetbox/backend-user/internal/service"
)

type Dependencies struct {
	Config config.Config
	DB     *gorm.DB
	Logger *slog.Logger
	Clock  service.Clock
	Hasher service.PasswordHasher
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
	if d.Hasher == nil {
		d.Hasher = security.NewBcryptHasher(security.DefaultBcryptCost)
	}

	tokens := security.NewJWTService(d.Config.JWTSecret, d.Config.JWTTTL)
	authService := service.NewAuthService(repository.NewUserRepository(d.DB), d.Hasher, tokens, d.Clock)

	return router.New(router.Config{
		Auth:           handler.NewAuthHandler(authService),
		System:         handler.NewSystemHandler(docs.OpenAPI, docs.SwaggerHTML),
		Authenticator:  authService,
		Logger:         d.Logger,
		AllowedOrigins: d.Config.CORSAllowedOrigins,
	})
}
