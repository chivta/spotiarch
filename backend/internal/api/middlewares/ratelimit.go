package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/chivta/spotiarch/internal/api/handlers"
	"github.com/chivta/spotiarch/internal/shared/domain"
)

type rateLimitCache interface {
	Allow(ctx context.Context, key string, limit, windowSeconds int) (bool, error)
}

const RateLimitContextKey = "ratelimit:"

func NewRateLimitMiddleware(cache rateLimitCache) *RateLimitMiddleware {
	return &RateLimitMiddleware{cache: cache}
}

type RateLimitMiddleware struct {
	cache rateLimitCache
}

func (m *RateLimitMiddleware) LimitRequests(limit, windowSeconds int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		allowed, err := m.cache.Allow(c.UserContext(), RateLimitContextKey+c.IP(), limit, windowSeconds)
		if err != nil {
			return c.Next()
		}
		if !allowed {
			return handlers.RespondWithError(c, domain.ErrTooManyRequests)
		}
		return c.Next()
	}
}
