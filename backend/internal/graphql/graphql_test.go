package graphql

import (
	"context"
	"testing"
	"time"

	"github.com/mathianasj/access-tracker/internal/models"
)

func TestAccessRequestModel(t *testing.T) {
	ar := &models.AccessRequest{
		ID:             "test-id",
		Requester:      "test-user",
		SystemResource: "test-system",
		AccessLevel:    models.AccessLevelRead,
		Justification:  "Test justification",
		Status:         models.StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if ar.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%s'", ar.ID)
	}

	if ar.Requester != "test-user" {
		t.Errorf("expected requester 'test-user', got '%s'", ar.Requester)
	}

	if ar.Status != models.StatusPending {
		t.Errorf("expected status 'pending', got '%s'", ar.Status)
	}
}

func TestAccessLevelConstants(t *testing.T) {
	if models.AccessLevelRead != "read" {
		t.Errorf("expected AccessLevelRead to be 'read', got '%s'", models.AccessLevelRead)
	}

	if models.AccessLevelWrite != "write" {
		t.Errorf("expected AccessLevelWrite to be 'write', got '%s'", models.AccessLevelWrite)
	}

	if models.AccessLevelAdmin != "admin" {
		t.Errorf("expected AccessLevelAdmin to be 'admin', got '%s'", models.AccessLevelAdmin)
	}
}

func TestStatusConstants(t *testing.T) {
	if models.StatusPending != "pending" {
		t.Errorf("expected StatusPending to be 'pending', got '%s'", models.StatusPending)
	}

	if models.StatusApproved != "approved" {
		t.Errorf("expected StatusApproved to be 'approved', got '%s'", models.StatusApproved)
	}

	if models.StatusDenied != "denied" {
		t.Errorf("expected StatusDenied to be 'denied', got '%s'", models.StatusDenied)
	}
}

func TestSchemaTypes(t *testing.T) {
	if accessRequestType == nil {
		t.Error("accessRequestType should not be nil")
	}

	if accessLevelEnum == nil {
		t.Error("accessLevelEnum should not be nil")
	}

	if statusEnum == nil {
		t.Error("statusEnum should not be nil")
	}

	if queryType == nil {
		t.Error("queryType should not be nil")
	}

	if mutationType == nil {
		t.Error("mutationType should not be nil")
	}

	if createAccessRequestInput == nil {
		t.Error("createAccessRequestInput should not be nil")
	}

	if updateAccessRequestStatusInput == nil {
		t.Error("updateAccessRequestStatusInput should not be nil")
	}

	if accessRequestFilterInput == nil {
		t.Error("accessRequestFilterInput should not be nil")
	}
}

func TestResolverCreation(t *testing.T) {
	r := NewResolver(nil)
	if r == nil {
		t.Error("NewResolver should not return nil")
	}
	if r.DB != nil {
		t.Error("NewResolver with nil DB should have nil DB field")
	}
	_ = context.Background()
}
