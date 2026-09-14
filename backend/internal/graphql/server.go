package graphql

import (
	"context"
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/mathianasj/access-tracker/internal/db"
)

func NewHandler(database *db.DB) *handler.Handler {
	h := handler.New(&handler.Config{
		Schema:   &Schema,
		Pretty:   true,
		GraphiQL: true,
	})

	return h
}

func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestID, ok := r.Context().Value("request_id").(string); ok {
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
