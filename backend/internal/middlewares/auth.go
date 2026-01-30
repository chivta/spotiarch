package middlewares

import (
	"github.com/chivta/spotiarch/internal/logger"
	"github.com/chivta/spotiarch/internal/services"
	"github.com/gofiber/fiber/v2"
)

func NewAuthMiddleware(service *services.AuthService, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		s: service,
		log: logger,
	}
}

type AuthMiddleware struct {
	s *services.AuthService
	log *logger.Logger
}
func (m *AuthMiddleware) New() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Cookies("token")
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		token, err := m.s.ParseToken(tokenStr)
		switch err {
		case nil:
			// Token is valid, proceed to the next handler
			c.Locals("userID", token.UserID)
			return c.Next()
		case services.ErrTokenInvalid:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		case services.ErrTokenExpired:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token has expired",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}
}
