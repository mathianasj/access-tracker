package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/mathianasj/access-tracker/internal/db"
	"github.com/mathianasj/access-tracker/internal/graphql"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
}

func main() {
	ctx := context.Background()

	database, err := db.New(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to connect to database")
		os.Exit(1)
	}
	defer database.Close()

	graphqlHandler := graphql.NewHandler(database)

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(requestIDMiddleware)
	r.Use(requestLogger)

	r.Get("/healthz", healthHandler(database))
	r.Handle("/graphql", graphql.ContextMiddleware(graphqlHandler))

	logger.Info().Msg("server starting on :8080")
	http.ListenAndServe(":8080", r)
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func healthHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		response := struct {
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
		}{
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := database.Ping(ctx); err != nil {
			response.Status = "unhealthy"
			logger.Error().Str("request_id", requestID).Str("component", "backend").Str("operation", "health_check").Err(err).Msg("health check failed")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response)
			return
		}

		response.Status = "healthy"
		logger.Info().Str("request_id", requestID).Str("component", "backend").Str("operation", "health_check").Msg("health check passed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
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

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := getRequestID(r.Context())

		mw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(w, r)

		logger.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", mw.status).
			Dur("duration", time.Since(start)).
			Str("remote_addr", r.RemoteAddr).
			Str("component", "backend").
			Str("operation", "http_request").
			Msg("request")
	})
}

func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		return requestID
	}
	return "unknown"
}
