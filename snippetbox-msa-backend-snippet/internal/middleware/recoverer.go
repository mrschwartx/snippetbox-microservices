package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"snippetbox/backend-snippet/internal/dto"
)

func Recoverer(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				log.Error("panic recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", r.Header.Get(RequestIDHeader),
				)

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(dto.NewErrorResponse(dto.CodeInternal, "Internal server error"))
			}()

			next.ServeHTTP(w, r)
		})
	}
}
