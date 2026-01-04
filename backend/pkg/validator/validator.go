package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new Validator
func New() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return v.errors.HasErrors()
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

// Required validates that a string is not empty
func (v *Validator) Required(field, value, message string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, message)
	}
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int, message string) {
	if len(value) < min {
		v.AddError(field, message)
	}
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int, message string) {
	if len(value) > max {
		v.AddError(field, message)
	}
}

// Email validates email format
func (v *Validator) Email(field, value, message string) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.AddError(field, message)
	}
}

// UUID validates UUID format
func (v *Validator) UUID(field, value, message string) {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(value) {
		v.AddError(field, message)
	}
}

// InRange validates that a number is within a range
func (v *Validator) InRange(field string, value, min, max float64, message string) {
	if value < min || value > max {
		v.AddError(field, message)
	}
}

// OneOf validates that a value is one of the allowed values
func (v *Validator) OneOf(field, value string, allowed []string, message string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.AddError(field, message)
}

// Password validates password strength
func (v *Validator) Password(field, value, message string) {
	// At least 8 characters, one uppercase, one lowercase, one digit
	if len(value) < 8 {
		v.AddError(field, message)
		return
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(value)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(value)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(value)

	if !hasUpper || !hasLower || !hasDigit {
		v.AddError(field, message)
	}
}
