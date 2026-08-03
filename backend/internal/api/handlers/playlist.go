package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/chivta/spotiarch/internal/api/services"
	"github.com/chivta/spotiarch/internal/shared/domain"
)

func NewPlaylistHandler(service *services.PlaylistService, validate *validator.Validate) *PlaylistHandler {
	return &PlaylistHandler{service: service, validate: validate}
}

type PlaylistHandler struct {
	service  *services.PlaylistService
	validate *validator.Validate
}

func (h *PlaylistHandler) ResolvePlaylist(c *fiber.Ctx) error {
	var dto domain.ResolvePlaylistDTO
	if err := c.BodyParser(&dto); err != nil {
		return RespondWithError(c, domain.ErrBadRequest)
	}
	if err := h.validate.Struct(dto); err != nil {
		if _, ok := err.(validator.ValidationErrors); !ok {
			return RespondWithError(c, err)
		}
		return RespondWithError(c, domain.ErrBadRequest)
	}
	anonID, userID, err := requestOwner(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	response, err := h.service.Resolve(c.UserContext(), anonID, userID, dto.URL)
	if err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(response)
}

func (h *PlaylistHandler) GetPending(c *fiber.Ctx) error {
	anonID, userID, err := requestOwner(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	response, err := h.service.GetPending(c.UserContext(), anonID, userID)
	if err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(response)
}

func (h *PlaylistHandler) IssueVerificationToken(c *fiber.Ctx) error {
	userID, err := requestUserID(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	response, err := h.service.IssueVerificationToken(c.UserContext(), userID)
	if err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(response)
}

func requestOwner(c *fiber.Ctx) (string, *int, error) {
	role, ok := c.Locals(domain.UserRoleKey).(domain.Role)
	if !ok {
		return "", nil, domain.ErrUnauthorized
	}
	if role == domain.RoleAnon {
		anonID, ok := c.Locals(domain.UserIDKey).(string)
		if !ok {
			return "", nil, domain.ErrUnauthorized
		}
		return anonID, nil, nil
	}
	userID, err := requestUserID(c)
	if err != nil {
		return "", nil, err
	}
	return "", &userID, nil
}
