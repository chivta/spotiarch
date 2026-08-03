package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Logger(skipPaths ...string) fiber.Handler {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skip[path] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		if _, ok := skip[c.Path()]; ok {
			return err
		}

		event := log.Info()
		if c.Response().StatusCode() >= fiber.StatusInternalServerError {
			event = log.Error()
		}
		event.
			Str("method", c.Method()).
			Str("path", c.OriginalURL()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Str("ip", c.IP()).
			Msg("request")
		return err
	}
}
