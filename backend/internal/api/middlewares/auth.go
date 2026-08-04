package middlewares

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/api/handlers"
	"github.com/chivta/spotiarch/internal/api/services"
	"github.com/chivta/spotiarch/internal/shared/domain"
)

type AuthMiddleware struct {
	authService   *services.AuthService
	secureCookies bool
}

func NewAuthMiddleware(authService *services.AuthService, secureCookies bool) *AuthMiddleware {
	return &AuthMiddleware{authService: authService, secureCookies: secureCookies}
}

func (m *AuthMiddleware) ParseAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtStr := c.Cookies(domain.CookieJWT)
		if jwtStr == "" {
			if err := m.issueAnonSession(c); err != nil {
				return err
			}
			return c.Next()
		}

		claims, err := m.authService.ParseJWT(jwtStr)
		if err == nil {
			c.Locals(domain.UserIDKey, claims.UserID)
			c.Locals(domain.UserRoleKey, claims.Role)
			return c.Next()
		}
		if !errors.Is(err, jwt.ErrTokenExpired) {
			m.clearCookies(c)
			return handlers.RespondWithError(c, domain.ErrUnauthorized)
		}
		if claims.Role == domain.RoleAnon {
			if err := m.issueAnonSession(c); err != nil {
				return err
			}
			return c.Next()
		}

		refreshToken := c.Cookies(domain.CookieRefreshToken)
		if refreshToken == "" {
			m.clearCookies(c)
			return handlers.RespondWithError(c, domain.ErrUnauthorized)
		}
		session, err := m.authService.ExchangeRefreshToken(c.UserContext(), jwtStr, refreshToken)
		if err != nil {
			m.clearCookies(c)
			return handlers.RespondWithError(c, domain.ErrUnauthorized)
		}
		m.setSessionCookies(c, session)
		c.Locals(domain.UserIDKey, session.UserID)
		c.Locals(domain.UserRoleKey, session.Role)
		return c.Next()
	}
}

func (m *AuthMiddleware) RequireUserRole() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(domain.UserRoleKey).(domain.Role)
		if !ok {
			log.Error().Msg("user role not found in request context")
			return handlers.RespondWithError(c, domain.ErrInternal)
		}
		if role != domain.RoleUser {
			return handlers.RespondWithError(c, domain.ErrForbidden)
		}
		return c.Next()
	}
}

func (m *AuthMiddleware) RequireAnonQuota(path string, limit int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(domain.UserRoleKey).(domain.Role)
		if !ok {
			log.Error().Msg("user role not found in request context")
			return handlers.RespondWithError(c, domain.ErrInternal)
		}
		if role != domain.RoleAnon {
			return c.Next()
		}
		anonID, ok := c.Locals(domain.UserIDKey).(string)
		if !ok {
			log.Error().Msg("anonymous user id not found in request context")
			return handlers.RespondWithError(c, domain.ErrInternal)
		}
		count, err := m.authService.IncrementAnonQuota(c.UserContext(), anonID, path)
		if err != nil {
			return handlers.RespondWithError(c, err)
		}
		if count > limit {
			return handlers.RespondWithError(c, domain.ErrQuotaExceeded)
		}
		return c.Next()
	}
}

func (m *AuthMiddleware) issueAnonSession(c *fiber.Ctx) error {
	session, err := m.authService.CreateAnonymousSession(c.UserContext())
	if err != nil {
		return handlers.RespondWithError(c, err)
	}
	c.Cookie(&fiber.Cookie{
		Name: domain.CookieJWT, Value: session.JWT, MaxAge: domain.AnonSessionCookieAge,
		Path: "/", HTTPOnly: true, Secure: m.secureCookies, SameSite: fiber.CookieSameSiteLaxMode,
	})
	c.Locals(domain.UserIDKey, session.UserID)
	c.Locals(domain.UserRoleKey, session.Role)
	return nil
}

func (m *AuthMiddleware) setSessionCookies(c *fiber.Ctx, session *services.Session) {
	c.Cookie(&fiber.Cookie{
		Name: domain.CookieJWT, Value: session.JWT, MaxAge: domain.JWTCookieAge,
		Path: "/", HTTPOnly: true, Secure: m.secureCookies, SameSite: fiber.CookieSameSiteLaxMode,
	})
	c.Cookie(&fiber.Cookie{
		Name: domain.CookieRefreshToken, Value: session.RefreshToken, MaxAge: domain.RefreshTokenCookieAge,
		Path: "/", HTTPOnly: true, Secure: m.secureCookies, SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func (m *AuthMiddleware) clearCookies(c *fiber.Ctx) {
	for _, name := range []string{domain.CookieJWT, domain.CookieRefreshToken} {
		c.Cookie(&fiber.Cookie{
			Name: name, Value: "", MaxAge: -1, Path: "/", HTTPOnly: true,
			Secure: m.secureCookies, SameSite: fiber.CookieSameSiteLaxMode,
		})
	}
}
