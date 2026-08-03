package handlers

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chivta/spotiarch/internal/api/services"
	"github.com/chivta/spotiarch/internal/shared/domain"
)

func NewAuthHandler(authService *services.AuthService, validate *validator.Validate, secureCookies bool) *AuthHandler {
	return &AuthHandler{authService: authService, validate: validate, secureCookies: secureCookies}
}

type AuthHandler struct {
	authService   *services.AuthService
	validate      *validator.Validate
	secureCookies bool
}

func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var dto domain.SignupDTO
	if err := c.BodyParser(&dto); err != nil {
		return RespondWithError(c, domain.ErrBadRequest)
	}
	if err := h.validate.Struct(dto); err != nil {
		if _, ok := err.(validator.ValidationErrors); !ok {
			return RespondWithError(c, err)
		}
		return RespondWithError(c, domain.ErrBadRequest)
	}
	session, err := h.authService.Signup(c.UserContext(), dto, anonymousID(c))
	if err != nil {
		return RespondWithError(c, err)
	}
	h.setSessionCookies(c, session)
	return c.JSON(fiber.Map{"message": "signup successful"})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var dto domain.LoginDTO
	if err := c.BodyParser(&dto); err != nil {
		return RespondWithError(c, domain.ErrBadRequest)
	}
	if err := h.validate.Struct(dto); err != nil {
		if _, ok := err.(validator.ValidationErrors); !ok {
			return RespondWithError(c, err)
		}
		return RespondWithError(c, domain.ErrBadRequest)
	}
	session, err := h.authService.Login(c.UserContext(), dto, anonymousID(c))
	if err != nil {
		return RespondWithError(c, err)
	}
	h.setSessionCookies(c, session)
	return c.JSON(fiber.Map{"message": "login successful"})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, err := requestUserID(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	if err := h.authService.Logout(c.UserContext(), userID); err != nil {
		return RespondWithError(c, err)
	}
	h.clearCookies(c)
	return c.JSON(fiber.Map{"message": "logout successful"})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		domain.UserIDKey: c.Locals(domain.UserIDKey), domain.UserRoleKey: c.Locals(domain.UserRoleKey),
	})
}

func (h *AuthHandler) setSessionCookies(c *fiber.Ctx, session *services.Session) {
	c.Cookie(&fiber.Cookie{
		Name: domain.CookieJWT, Value: session.JWT, MaxAge: domain.JWTCookieAge,
		Path: "/", HTTPOnly: true, Secure: h.secureCookies, SameSite: fiber.CookieSameSiteLaxMode,
	})
	c.Cookie(&fiber.Cookie{
		Name: domain.CookieRefreshToken, Value: session.RefreshToken, MaxAge: domain.RefreshTokenCookieAge,
		Path: "/", HTTPOnly: true, Secure: h.secureCookies, SameSite: fiber.CookieSameSiteLaxMode,
	})
}

func (h *AuthHandler) clearCookies(c *fiber.Ctx) {
	for _, name := range []string{domain.CookieJWT, domain.CookieRefreshToken} {
		c.Cookie(&fiber.Cookie{
			Name: name, Value: "", MaxAge: -1, Path: "/", HTTPOnly: true,
			Secure: h.secureCookies, SameSite: fiber.CookieSameSiteLaxMode,
		})
	}
}

func anonymousID(c *fiber.Ctx) string {
	if role, _ := c.Locals(domain.UserRoleKey).(domain.Role); role != domain.RoleAnon {
		return ""
	}
	id, _ := c.Locals(domain.UserIDKey).(string)
	return id
}

func requestUserID(c *fiber.Ctx) (int, error) {
	id, ok := c.Locals(domain.UserIDKey).(string)
	if !ok {
		return 0, domain.ErrUnauthorized
	}
	parsed, err := strconv.Atoi(id)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}
	return parsed, nil
}
