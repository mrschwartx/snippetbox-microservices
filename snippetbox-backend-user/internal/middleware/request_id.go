package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	RequestIDHeader    = "X-Request-ID"
	maxRequestIDLength = 64
	generatedIDBytes   = 8
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(RequestIDHeader)
			if id == "" || len(id) > maxRequestIDLength {
				id = newRequestID()
			}

			r.Header.Set(RequestIDHeader, id)
			w.Header().Set(RequestIDHeader, id)
			next.ServeHTTP(w, r)
		})
	}
}

func newRequestID() string {
	buf := make([]byte, generatedIDBytes)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
