package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AuditLog struct {
	ID              string    `json:"id"`
	AccessRequestID string    `json:"accessRequestId"`
	Action          string    `json:"action"`
	OldValue        string    `json:"oldValue,omitempty"`
	NewValue        string    `json:"newValue,omitempty"`
	ChangedBy       string    `json:"changedBy"`
	ChangedAt       time.Time `json:"changedAt"`
}
