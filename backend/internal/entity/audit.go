package entity

import (
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	UserID       *uuid.UUID      `json:"user_id,omitempty" db:"user_id"`
	Action       string          `json:"action" db:"action"`
	ResourceType string          `json:"resource_type" db:"resource_type"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty" db:"resource_id"`
	OldValue     json.RawMessage `json:"old_value,omitempty" db:"old_value"`
	NewValue     json.RawMessage `json:"new_value,omitempty" db:"new_value"`
	IPAddress    net.IP          `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string          `json:"user_agent,omitempty" db:"user_agent"`
	RequestID    string          `json:"request_id,omitempty" db:"request_id"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// Audit action constants
const (
	AuditActionCreate = "CREATE"
	AuditActionRead   = "READ"
	AuditActionUpdate = "UPDATE"
	AuditActionDelete = "DELETE"
	AuditActionLogin  = "LOGIN"
	AuditActionLogout = "LOGOUT"
)

// Resource type constants
const (
	ResourceUser    = "user"
	ResourcePatient = "patient"
	ResourceSurvey  = "survey"
	ResourceTherapy = "therapy"
	ResourceDrug    = "drug"
)

// AuditLogCreate represents data for creating an audit log entry
type AuditLogCreate struct {
	UserID       *uuid.UUID  `json:"user_id,omitempty"`
	Action       string      `json:"action"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	OldValue     interface{} `json:"old_value,omitempty"`
	NewValue     interface{} `json:"new_value,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	RequestID    string      `json:"request_id,omitempty"`
}

// AuditLogFilter represents filter options for audit logs
type AuditLogFilter struct {
	UserID       *uuid.UUID `query:"user_id"`
	Action       string     `query:"action"`
	ResourceType string     `query:"resource_type"`
	ResourceID   *uuid.UUID `query:"resource_id"`
	DateFrom     *time.Time `query:"date_from"`
	DateTo       *time.Time `query:"date_to"`
	Page         int        `query:"page"`
	PerPage      int        `query:"per_page"`
}
