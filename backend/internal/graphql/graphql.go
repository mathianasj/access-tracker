package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/mathianasj/access-tracker/internal/db"
	"github.com/mathianasj/access-tracker/internal/models"
)

var accessLevelEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "AccessLevel",
	Values: graphql.EnumValueConfigMap{
		"READ":  &graphql.EnumValueConfig{Value: models.AccessLevelRead},
		"WRITE": &graphql.EnumValueConfig{Value: models.AccessLevelWrite},
		"ADMIN": &graphql.EnumValueConfig{Value: models.AccessLevelAdmin},
	},
})

var statusEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "Status",
	Values: graphql.EnumValueConfigMap{
		"PENDING":  &graphql.EnumValueConfig{Value: models.StatusPending},
		"APPROVED": &graphql.EnumValueConfig{Value: models.StatusApproved},
		"DENIED":   &graphql.EnumValueConfig{Value: models.StatusDenied},
	},
})

var accessRequestType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AccessRequest",
	Fields: graphql.Fields{
		"id":             &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"requester":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"systemResource": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"accessLevel":    &graphql.Field{Type: graphql.NewNonNull(accessLevelEnum)},
		"justification":  &graphql.Field{Type: graphql.String},
		"status":         &graphql.Field{Type: graphql.NewNonNull(statusEnum)},
		"createdAt":      &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"updatedAt":      &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
	},
})

var createAccessRequestInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateAccessRequestInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"requester":      &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"systemResource": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"accessLevel":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(accessLevelEnum)},
		"justification":  &graphql.InputObjectFieldConfig{Type: graphql.String},
	},
})

var updateAccessRequestStatusInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateAccessRequestStatusInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id":     &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.ID)},
		"status": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(statusEnum)},
	},
})

var accessRequestFilterInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AccessRequestFilter",
	Fields: graphql.InputObjectConfigFieldMap{
		"status":    &graphql.InputObjectFieldConfig{Type: statusEnum},
		"requester": &graphql.InputObjectFieldConfig{Type: graphql.String},
	},
})

type Resolver struct {
	DB *db.DB
}

func NewResolver(database *db.DB) *Resolver {
	return &Resolver{DB: database}
}

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"accessRequest": &graphql.Field{
			Type:        accessRequestType,
			Description: "Get an access request by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := p.Args["id"].(string)
				return p.Context.Value("resolver").(*Resolver).getAccessRequest(p.Context, id)
			},
		},
		"accessRequests": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(accessRequestType))),
			Description: "Get all access requests with optional filters",
			Args: graphql.FieldConfigArgument{
				"filter": &graphql.ArgumentConfig{Type: accessRequestFilterInput},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				filter, _ := p.Args["filter"].(map[string]interface{})
				return p.Context.Value("resolver").(*Resolver).getAccessRequests(p.Context, filter)
			},
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"createAccessRequest": &graphql.Field{
			Type:        accessRequestType,
			Description: "Create a new access request",
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createAccessRequestInput)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				input := p.Args["input"].(map[string]interface{})
				return p.Context.Value("resolver").(*Resolver).createAccessRequest(p.Context, input)
			},
		},
		"updateAccessRequestStatus": &graphql.Field{
			Type:        accessRequestType,
			Description: "Update the status of an access request",
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(updateAccessRequestStatusInput)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				input := p.Args["input"].(map[string]interface{})
				return p.Context.Value("resolver").(*Resolver).updateAccessRequestStatus(p.Context, input)
			},
		},
	},
})

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    queryType,
	Mutation: mutationType,
})

func (r *Resolver) getAccessRequest(ctx context.Context, id string) (*models.AccessRequest, error) {
	ar := &models.AccessRequest{}
	err := r.DB.Pool.QueryRow(ctx, `
		SELECT id, requester, system_resource, access_level, justification, status, created_at, updated_at
		FROM access_requests
		WHERE id = $1
	`, id).Scan(
		&ar.ID, &ar.Requester, &ar.SystemResource, &ar.AccessLevel, &ar.Justification, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ar, nil
}

func (r *Resolver) getAccessRequests(ctx context.Context, filter map[string]interface{}) ([]*models.AccessRequest, error) {
	query := `
		SELECT id, requester, system_resource, access_level, justification, status, created_at, updated_at
		FROM access_requests
		WHERE 1=1
	`
	var args []interface{}

	if filter != nil {
		if status, ok := filter["status"]; ok && status != nil {
			query += " AND status = $1"
			args = append(args, status.(models.Status))
		}
		if requester, ok := filter["requester"]; ok && requester != nil {
			if len(args) == 0 {
				query += " AND requester = $1"
			} else {
				query += " AND requester = $2"
			}
			args = append(args, requester.(string))
		}
	}

	rows, err := r.DB.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.AccessRequest
	for rows.Next() {
		ar := &models.AccessRequest{}
		err := rows.Scan(
			&ar.ID, &ar.Requester, &ar.SystemResource, &ar.AccessLevel, &ar.Justification, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, ar)
	}

	return results, nil
}

func (r *Resolver) createAccessRequest(ctx context.Context, input map[string]interface{}) (*models.AccessRequest, error) {
	ar := &models.AccessRequest{
		Requester:      input["requester"].(string),
		SystemResource: input["systemResource"].(string),
		AccessLevel:    input["accessLevel"].(models.AccessLevel),
		Status:         models.StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if justification, ok := input["justification"]; ok && justification != nil {
		ar.Justification = justification.(string)
	}

	err := r.DB.Pool.QueryRow(ctx, `
		INSERT INTO access_requests (requester, system_resource, access_level, justification, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, ar.Requester, ar.SystemResource, ar.AccessLevel, ar.Justification, ar.Status, ar.CreatedAt, ar.UpdatedAt).Scan(&ar.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create access request: %w", err)
	}

	return ar, nil
}

func (r *Resolver) updateAccessRequestStatus(ctx context.Context, input map[string]interface{}) (*models.AccessRequest, error) {
	id := input["id"].(string)
	status := input["status"].(models.Status)

	ar := &models.AccessRequest{}
	err := r.DB.Pool.QueryRow(ctx, `
		UPDATE access_requests
		SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, requester, system_resource, access_level, justification, status, created_at, updated_at
	`, status, time.Now(), id).Scan(
		&ar.ID, &ar.Requester, &ar.SystemResource, &ar.AccessLevel, &ar.Justification, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update access request status: %w", err)
	}

	return ar, nil
}
