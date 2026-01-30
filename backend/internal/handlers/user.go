package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/chivta/spotiarch/internal/models"
	"github.com/chivta/spotiarch/internal/services"
)

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		s: service,
	}
}

type UserHandler struct {
	s *services.UserService
}

func (h *UserHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := h.s.GetUserByID(userID)
	switch err {
	case services.ErrUserNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	case services.ErrInternal:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	userResponse := &models.UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}

	return c.Status(fiber.StatusOK).JSON(userResponse)
}
