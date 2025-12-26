package middleware

import (
	"fmt"

	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/api/models"
	"github.com/gofiber/fiber/v2"
)

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		fmt.Printf("[ERROR] %s %s: %v\n", c.Method(), c.Path(), err)

		return c.Status(code).JSON(models.ErrorResponse{
			Error:   getErrorTitle(code),
			Message: err.Error(),
			Code:    code,
		})
	}
}

func getErrorTitle(code int) string {
	switch code {
	case fiber.StatusBadRequest:
		return "Bad Request"
	case fiber.StatusUnauthorized:
		return "Unauthorized"
	case fiber.StatusForbidden:
		return "Forbidden"
	case fiber.StatusNotFound:
		return "Not Found"
	case fiber.StatusMethodNotAllowed:
		return "Method Not Allowed"
	case fiber.StatusRequestTimeout:
		return "Request Timeout"
	case fiber.StatusTooManyRequests:
		return "Too Many Requests"
	case fiber.StatusInternalServerError:
		return "Internal Server Error"
	case fiber.StatusBadGateway:
		return "Bad Gateway"
	case fiber.StatusServiceUnavailable:
		return "Service Unavailable"
	default:
		return "Error"
	}
}
