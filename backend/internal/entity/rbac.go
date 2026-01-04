package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user role in the system
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission represents an action permission
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Common role names
const (
	RoleAdmin   = "admin"
	RoleDoctor  = "doctor"
	RoleNurse   = "nurse"
	RolePatient = "patient"
)

// Common permission names
const (
	// User permissions
	PermUsersRead   = "users:read"
	PermUsersWrite  = "users:write"
	PermUsersDelete = "users:delete"

	// Patient permissions
	PermPatientsRead    = "patients:read"
	PermPatientsReadOwn = "patients:read_own"
	PermPatientsWrite   = "patients:write"
	PermPatientsDelete  = "patients:delete"

	// Survey permissions
	PermSurveysRead   = "surveys:read"
	PermSurveysSubmit = "surveys:submit"
	PermSurveysReview = "surveys:review"
	PermSurveysManage = "surveys:manage"

	// Therapy permissions
	PermTherapyRead     = "therapy:read"
	PermTherapyReadOwn  = "therapy:read_own"
	PermTherapyWrite    = "therapy:write"
	PermTherapySchedule = "therapy:schedule"

	// Drug permissions
	PermDrugsRead  = "drugs:read"
	PermDrugsWrite = "drugs:write"

	// Admin
	PermAdminFull = "admin:full"
)

// Default role UUIDs (matching migration seed data)
var (
	AdminRoleID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	DoctorRoleID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	NurseRoleID   = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	PatientRoleID = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

// GetRoleIDByName returns the UUID for a role name
func GetRoleIDByName(name string) uuid.UUID {
	switch name {
	case RoleAdmin:
		return AdminRoleID
	case RoleDoctor:
		return DoctorRoleID
	case RoleNurse:
		return NurseRoleID
	case RolePatient:
		return PatientRoleID
	default:
		return uuid.Nil
	}
}
