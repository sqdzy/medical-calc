package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	FirstName     string     `json:"first_name,omitempty" db:"first_name"`
	LastName      string     `json:"last_name,omitempty" db:"last_name"`
	Phone         string     `json:"phone,omitempty" db:"phone"`
	RoleID        uuid.UUID  `json:"role_id" db:"role_id"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`

	// Joined fields
	Role        *Role    `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// UserCreate represents data for creating a new user
type UserCreate struct {
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"password" validate:"required,min=8"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	RoleID    uuid.UUID `json:"role_id"`
}

// UserUpdate represents data for updating a user
type UserUpdate struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission || p == PermAdminFull {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if user has any of the specified permissions
func (u *User) HasAnyPermission(permissions ...string) bool {
	for _, p := range permissions {
		if u.HasPermission(p) {
			return true
		}
	}
	return false
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.Role != nil && u.Role.Name == RoleAdmin
}

// IsDoctor checks if user is a doctor
func (u *User) IsDoctor() bool {
	return u.Role != nil && u.Role.Name == RoleDoctor
}

// IsPatient checks if user is a patient
func (u *User) IsPatient() bool {
	return u.Role != nil && u.Role.Name == RolePatient
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return u.FirstName + " " + u.LastName
}

// Sanitize removes sensitive data from user
func (u *User) Sanitize() {
	u.PasswordHash = ""
}
