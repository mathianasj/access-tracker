package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		mw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(mw, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", mw.status).
			Dur("duration", time.Since(start)).
			Str("remote_addr", r.RemoteAddr).
			Msg("request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
