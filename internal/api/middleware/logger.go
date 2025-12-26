package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		fmt.Printf("[%s] %d %s %s - %s\n",
			start.Format("2006-01-02 15:04:05"),
			statusCode,
			c.Method(),
			c.Path(),
			duration,
		)

		return err
	}
}
