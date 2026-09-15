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

func ContextMiddleware(resolver *Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if requestID, ok := ctx.Value("request_id").(string); ok {
				ctx = context.WithValue(ctx, "request_id", requestID)
			}

			if username, ok := ctx.Value("username").(string); ok {
				ctx = context.WithValue(ctx, "username", username)
			}

			ctx = context.WithValue(ctx, "resolver", resolver)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
