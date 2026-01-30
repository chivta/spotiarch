package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/chivta/spotiarch/internal/models"
	"github.com/chivta/spotiarch/internal/services"
)

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{s: service}
}

type AuthHandler struct {
	s *services.AuthService
}

func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	signUpReq := &models.SignUpDTO{}
	err := c.BodyParser(signUpReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	token, err := h.s.SignUp(signUpReq.Email, signUpReq.Password)

	switch err {
	case nil: // all good
	case services.ErrEmailExists:
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User already exists",
		})
	case services.ErrInternal:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	default:
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Service unavailable",
		})
	}

	// Set token as HttpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours in seconds
		Secure:   false,        // TODO: set to true in production
		HTTPOnly: true,         // Not accessible via JavaScript
		SameSite: "Lax",        // CSRF protection
	})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	loginReq := &models.LoginDTO{}
	err := c.BodyParser(loginReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	token, err := h.s.Login(loginReq.Email, loginReq.Password)

	switch err {
	case nil: // all good
	case services.ErrInvalidCredentials:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	case services.ErrInternal:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	default:
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Service unavailable",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours in seconds
		Secure:   false,        // TODO: set to true in production
		HTTPOnly: true,         // Not accessible via JavaScript
		SameSite: "Lax",        // CSRF protection
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear the token cookie by setting it with negative max age
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   false, // TODO: set to true in production
		HTTPOnly: true,
		SameSite: "Lax",
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout successful",
	})
}
