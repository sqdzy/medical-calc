package v1

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/medical-app/backend/pkg/response"
	"github.com/medical-app/backend/pkg/validator"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	// Fiber error
	var e *fiber.Error
	if errors.As(err, &e) {
		return response.Error(c, e.Code, "HTTP_ERROR", e.Message)
	}

	// Validation errors
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return response.ValidationError(c, ve.Error())
	}

	return response.InternalError(c, "Internal server error")
}
