package middleware

import (
	"net/http"
	"slices"
)

func CORS(allowedOrigins []string) Middleware {
	allowAll := slices.Contains(allowedOrigins, "*")
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			header := w.Header()
			header.Add("Vary", "Origin")

			_, listed := allowed[origin]
			permitted := listed || allowAll
			if permitted {
				if allowAll {
					header.Set("Access-Control-Allow-Origin", "*")
				} else {
					header.Set("Access-Control-Allow-Origin", origin)
				}
				header.Set("Access-Control-Expose-Headers", RequestIDHeader+", Location")
			}

			if isPreflight(r) {
				if permitted {
					header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
					header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
					header.Set("Access-Control-Max-Age", "600")
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}
