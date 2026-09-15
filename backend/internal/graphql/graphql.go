package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/google/uuid"
	"github.com/mathianasj/access-tracker/internal/auth"
	"github.com/mathianasj/access-tracker/internal/db"
	"github.com/mathianasj/access-tracker/internal/models"
	"github.com/rs/zerolog/log"
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
		"requestId":      &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"requester":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"systemResource": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"accessLevel":    &graphql.Field{Type: graphql.NewNonNull(accessLevelEnum)},
		"justification":  &graphql.Field{Type: graphql.String},
		"status":         &graphql.Field{Type: graphql.NewNonNull(statusEnum)},
		"createdAt":      &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"updatedAt":      &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
	},
})

var auditLogType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuditLog",
	Fields: graphql.Fields{
		"id":              &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"accessRequestId": &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"action":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"oldValue":        &graphql.Field{Type: graphql.String},
		"newValue":        &graphql.Field{Type: graphql.String},
		"changedBy":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"changedAt":       &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
	},
})

var authPayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthPayload",
	Fields: graphql.Fields{
		"token":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"username": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
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

var loginInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "LoginInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"username": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"password": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
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
				requestID := getRequestID(p.Context)
				id := p.Args["id"].(string)
				log.Info().Str("request_id", requestID).Str("operation", "getAccessRequest").Str("component", "graphql").Msg("resolving access request")
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
				requestID := getRequestID(p.Context)
				filter, _ := p.Args["filter"].(map[string]interface{})
				log.Info().Str("request_id", requestID).Str("operation", "getAccessRequests").Str("component", "graphql").Msg("listing access requests")
				return p.Context.Value("resolver").(*Resolver).getAccessRequests(p.Context, filter)
			},
		},
		"auditLogs": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(auditLogType))),
			Description: "Get audit logs for an access request",
			Args: graphql.FieldConfigArgument{
				"accessRequestId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				requestID := getRequestID(p.Context)
				accessRequestID := p.Args["accessRequestId"].(string)
				log.Info().Str("request_id", requestID).Str("operation", "getAuditLogs").Str("component", "graphql").Msg("fetching audit logs")
				return p.Context.Value("resolver").(*Resolver).getAuditLogs(p.Context, accessRequestID)
			},
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"login": &graphql.Field{
			Type:        authPayloadType,
			Description: "Login with username and password",
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(loginInput)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				input := p.Args["input"].(map[string]interface{})
				username := input["username"].(string)
				password := input["password"].(string)
				log.Info().Str("operation", "login").Str("username", username).Str("component", "graphql").Msg("login attempt")
				return p.Context.Value("resolver").(*Resolver).login(p.Context, username, password)
			},
		},
		"createAccessRequest": &graphql.Field{
			Type:        accessRequestType,
			Description: "Create a new access request",
			Args: graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createAccessRequestInput)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				input := p.Args["input"].(map[string]interface{})
				requestID := getRequestID(p.Context)
				username := auth.GetUsernameFromContext(p.Context)
				log.Info().Str("request_id", requestID).Str("operation", "createAccessRequest").Str("component", "graphql").Str("username", username).Msg("creating access request")
				return p.Context.Value("resolver").(*Resolver).createAccessRequest(p.Context, input, username)
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
				requestID := getRequestID(p.Context)
				username := auth.GetUsernameFromContext(p.Context)
				log.Info().Str("request_id", requestID).Str("operation", "updateAccessRequestStatus").Str("component", "graphql").Str("username", username).Msg("updating access request status")
				return p.Context.Value("resolver").(*Resolver).updateAccessRequestStatus(p.Context, input, username)
			},
		},
	},
})

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    queryType,
	Mutation: mutationType,
})

func (r *Resolver) getAccessRequest(ctx context.Context, id string) (*models.AccessRequest, error) {
	requestID := getRequestID(ctx)
	log.Debug().Str("request_id", requestID).Str("operation", "getAccessRequest").Str("component", "db").Msg("querying database")

	ar := &models.AccessRequest{}
	err := r.DB.Pool.QueryRow(ctx, `
		SELECT id, request_id, requester, system_resource, access_level, justification, status, created_at, updated_at
		FROM access_requests
		WHERE id = $1
	`, id).Scan(
		&ar.ID, &ar.RequestID, &ar.Requester, &ar.SystemResource, &ar.AccessLevel, &ar.Justification, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ar, nil
}

func (r *Resolver) getAccessRequests(ctx context.Context, filter map[string]interface{}) ([]*models.AccessRequest, error) {
	requestID := getRequestID(ctx)
	log.Debug().Str("request_id", requestID).Str("operation", "getAccessRequests").Str("component", "db").Msg("querying database")

	query := `
		SELECT id, request_id, requester, system_resource, access_level, justification, status, created_at, updated_at
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
			&ar.ID, &ar.RequestID, &ar.Requester, &ar.SystemResource, &ar.AccessLevel, &ar.Justification, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, ar)
	}

	return results, nil
}

func (r *Resolver) login(ctx context.Context, username, password string) (map[string]interface{}, error) {
	var user models.User
	err := r.DB.Pool.QueryRow(ctx, `
		SELECT id, username, password_hash FROM users WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		log.Warn().Str("username", username).Msg("login failed - user not found")
		return nil, auth.ErrInvalidCredentials
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		log.Warn().Str("username", username).Msg("login failed - invalid password")
		return nil, auth.ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.Info().Str("username", username).Msg("login successful")
	return map[string]interface{}{
		"token":    token,
		"username": username,
	}, nil
}

func (r *Resolver) createAccessRequest(ctx context.Context, input map[string]interface{}, createdBy string) (*models.AccessRequest, error) {
	requestID := getRequestID(ctx)
	log.Debug().Str("request_id", requestID).Str("operation", "createAccessRequest").Str("component", "db").Msg("inserting into database")

	ar := &models.AccessRequest{
		RequestID:      requestID,
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
		INSERT INTO access_requests (request_id, requester, system_resource, access_level, justification, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, ar.RequestID, ar.Requester, ar.SystemResource, ar.AccessLevel, ar.Justification, ar.Status, ar.CreatedAt, ar.UpdatedAt).Scan(&ar.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create access request: %w", err)
	}

	_, err = r.DB.Pool.Exec(ctx, `
		INSERT INTO audit_logs (access_request_id, action, new_value, changed_by, changed_at)
		VALUES ($1, $2, $3, $4, $5)
	`, ar.ID, "CREATED", string(ar.Status), createdBy, time.Now())

	if err != nil {
		log.Warn().Err(err).Str("request_id", requestID).Msg("failed to create audit log for access request creation")
	}

	return ar, nil
}

func (r *Resolver) updateAccessRequestStatus(ctx context.Context, input map[string]interface{}, changedBy string) (*models.AccessRequest, error) {
	requestID := getRequestID(ctx)
	log.Debug().Str("request_id", requestID).Str("operation", "updateAccessRequestStatus").Str("component", "db").Msg("updating database")

	id := input["id"].(string)
	newStatus := input["status"].(models.Status)

	var oldStatus models.Status
	err := r.DB.Pool.QueryRow(ctx, `SELECT status FROM access_requests WHERE id = $1`, id).Scan(&oldStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to get current status: %w", err)
	}

	ar := &models.AccessRequest{}
	err = r.DB.Pool.QueryRow(ctx, `
		UPDATE access_requests
		SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, request_id, requester, system_resource, access_level, justification, status, created_at, updated_at
	`, newStatus, time.Now(), id).Scan(
		&ar.ID, &ar.RequestID, &ar.Requester, &ar.SystemResource, &ar.AccessLevel, &ar.Justification, &ar.Status, &ar.CreatedAt, &ar.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update access request status: %w", err)
	}

	_, err = r.DB.Pool.Exec(ctx, `
		INSERT INTO audit_logs (access_request_id, action, old_value, new_value, changed_by, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, ar.ID, "STATUS_CHANGED", string(oldStatus), string(newStatus), changedBy, time.Now())

	if err != nil {
		log.Warn().Err(err).Str("request_id", requestID).Msg("failed to create audit log for status change")
	}

	return ar, nil
}

func (r *Resolver) getAuditLogs(ctx context.Context, accessRequestID string) ([]*models.AuditLog, error) {
	requestID := getRequestID(ctx)
	log.Debug().Str("request_id", requestID).Str("operation", "getAuditLogs").Str("accessRequestId", accessRequestID).Str("component", "db").Msg("querying audit logs")

	rows, err := r.DB.Pool.Query(ctx, `
		SELECT id, access_request_id, action, old_value, new_value, changed_by, changed_at
		FROM audit_logs
		WHERE access_request_id = $1
		ORDER BY changed_at ASC
	`, accessRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.AuditLog
	for rows.Next() {
		al := &models.AuditLog{}
		var oldValue, newValue *string
		err := rows.Scan(&al.ID, &al.AccessRequestID, &al.Action, &oldValue, &newValue, &al.ChangedBy, &al.ChangedAt)
		if err != nil {
			return nil, err
		}
		if oldValue != nil {
			al.OldValue = *oldValue
		}
		if newValue != nil {
			al.NewValue = *newValue
		}
		results = append(results, al)
	}

	return results, nil
}

func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		return requestID
	}
	return uuid.New().String()
}
