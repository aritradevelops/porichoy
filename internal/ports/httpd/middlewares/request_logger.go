package middlewares

import (
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger.Info().Msgf("Request: %s %s", c.Method(), c.OriginalURL())
		// logger.Info().Msgf("Payload: %s", c.Body())
		// c.Response().StatusCode() this is giving wrong data
		logger.Info().Msgf("Response: %s %d", c.Response().Body(), c.Response().StatusCode())
		return c.Next()
	}
}
