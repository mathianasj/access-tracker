package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/mathianasj/access-tracker/internal/auth"
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
	resolver := graphql.NewResolver(database)

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(requestIDMiddleware)
	r.Use(requestLogger)
	r.Use(authMiddleware)

	r.Get("/healthz", healthHandler(database))
	r.Get("/me", meHandler())
	r.Get("/token", tokenHandler())
	r.Post("/logout", logoutHandler())
	r.Handle("/graphql", graphql.ContextMiddleware(resolver)(graphqlHandler))

	r.Get("/oauth/github", githubOAuthHandler(database))
	r.Get("/oauth/github/callback", githubOAuthCallbackHandler(database))

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

func meHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username")
		if username == nil || username == "anonymous" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"username": username.(string)})
	}
}

func tokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username")
		if username == nil || username == "anonymous" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
			return
		}

		tokenString, err := auth.GenerateToken(username.(string))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "username": username.(string)})
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"success": "true"})
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

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString := ""

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				tokenString = ""
			}
		}

		if tokenString == "" {
			if cookie, err := r.Cookie("auth_token"); err == nil {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			ctx = context.WithValue(ctx, "username", "anonymous")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			logger.Warn().Err(err).Str("operation", "auth").Msg("invalid token")
			ctx = context.WithValue(ctx, "username", "anonymous")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		ctx = context.WithValue(ctx, "username", claims.Username)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		return requestID
	}
	return "unknown"
}

func githubOAuthHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r.Context())
		state := auth.GenerateOAuthState()

		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   600,
		})

		url := auth.GitHubOAuth.GenerateAuthURL(state)
		logger.Info().Str("request_id", requestID).Str("redirect_url", url).Msg("oauth_github_initiated")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func githubOAuthCallbackHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r.Context())
		ctx := r.Context()

		stateCookie, err := r.Cookie("oauth_state")
		if err != nil || stateCookie == nil {
			logger.Warn().Str("request_id", requestID).Msg("oauth_state_cookie_missing")
			http.Redirect(w, r, "/?error=oauth_state_missing", http.StatusTemporaryRedirect)
			return
		}

		state := r.URL.Query().Get("state")
		if state != stateCookie.Value {
			logger.Warn().Str("request_id", requestID).Msg("oauth_state_mismatch")
			http.Redirect(w, r, "/?error=oauth_state_mismatch", http.StatusTemporaryRedirect)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			logger.Warn().Str("request_id", requestID).Msg("oauth_code_missing")
			http.Redirect(w, r, "/?error=oauth_code_missing", http.StatusTemporaryRedirect)
			return
		}

		token, err := auth.GitHubOAuth.ExchangeCode(ctx, code)
		if err != nil {
			auth.LogOAuthAttempt("github", "", false, err)
			logger.Error().Str("request_id", requestID).Err(err).Msg("oauth_token_exchange_failed")
			http.Redirect(w, r, "/?error=token_exchange_failed", http.StatusTemporaryRedirect)
			return
		}

		githubUser, err := auth.GitHubOAuth.GetUser(ctx, token)
		if err != nil {
			auth.LogOAuthAttempt("github", "", false, err)
			logger.Error().Str("request_id", requestID).Err(err).Msg("oauth_get_user_failed")
			http.Redirect(w, r, "/?error=get_user_failed", http.StatusTemporaryRedirect)
			return
		}

		username := githubUser.Login
		if githubUser.Email != "" {
			username = githubUser.Email
		}

		jwtToken, err := auth.GenerateToken(username)
		if err != nil {
			auth.LogOAuthAttempt("github", username, false, err)
			logger.Error().Str("request_id", requestID).Err(err).Msg("jwt_generation_failed")
			http.Redirect(w, r, "/?error=jwt_generation_failed", http.StatusTemporaryRedirect)
			return
		}

		err = database.LinkOAuthUser(ctx, "github", fmt.Sprintf("%d", githubUser.ID), username, string(token.AccessToken))
		if err != nil {
			logger.Warn().Str("request_id", requestID).Err(err).Msg("failed to link oauth user, continuing anyway")
		}

		auth.LogOAuthAttempt("github", username, true, nil)

		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    jwtToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   86400 * 7,
		})

		http.SetCookie(w, &http.Cookie{
			Name:   "oauth_state",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		})

		http.Redirect(w, r, "/?oauth=success", http.StatusTemporaryRedirect)
	}
}
