package graphql

import (
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
