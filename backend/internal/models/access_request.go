package models

import (
	"time"
)

type AccessLevel string

const (
	AccessLevelRead  AccessLevel = "read"
	AccessLevelWrite AccessLevel = "write"
	AccessLevelAdmin AccessLevel = "admin"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusDenied  Status = "denied"
)

type AccessRequest struct {
	ID             string      `json:"id"`
	RequestID      string      `json:"requestId"`
	Requester      string      `json:"requester"`
	SystemResource string      `json:"systemResource"`
	AccessLevel    AccessLevel `json:"accessLevel"`
	Justification  string      `json:"justification"`
	Status         Status      `json:"status"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}
