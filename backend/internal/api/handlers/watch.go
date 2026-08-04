package handlers

import (
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/chivta/spotiarch/internal/api/services"
	"github.com/chivta/spotiarch/internal/shared/domain"
)

func NewWatchHandler(service *services.ArchiveService) *WatchHandler {
	return &WatchHandler{service: service}
}

type WatchHandler struct {
	service *services.ArchiveService
}

func (h *WatchHandler) CreateWatch(c *fiber.Ctx) error {
	userID, err := requestUserID(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	response, err := h.service.CreateWatch(c.UserContext(), userID)
	if err != nil {
		return RespondWithError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h *WatchHandler) ListWatches(c *fiber.Ctx) error {
	userID, err := requestUserID(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	response, err := h.service.ListWatches(c.UserContext(), userID)
	if err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(response)
}

func (h *WatchHandler) GetWatch(c *fiber.Ctx) error {
	userID, watchID, err := requestWatchIDs(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	response, err := h.service.GetWatch(c.UserContext(), userID, watchID)
	if err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(response)
}

func (h *WatchHandler) DeleteWatch(c *fiber.Ctx) error {
	userID, watchID, err := requestWatchIDs(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	if err := h.service.DeleteWatch(c.UserContext(), userID, watchID); err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(fiber.Map{"message": "watch deleted"})
}

func (h *WatchHandler) ListTracks(c *fiber.Ctx) error {
	userID, watchID, err := requestWatchIDs(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	offset, limit, removedOnly, err := trackQuery(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	response, err := h.service.ListTracks(c.UserContext(), userID, watchID, removedOnly, offset, limit)
	if err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(response)
}

func (h *WatchHandler) DeleteTrack(c *fiber.Ctx) error {
	userID, watchID, err := requestWatchIDs(c)
	if err != nil {
		return RespondWithError(c, err)
	}
	uri, err := url.PathUnescape(c.Params("uri"))
	if err != nil || uri == "" {
		return RespondWithError(c, domain.ErrBadRequest)
	}
	if err := h.service.DeleteTrack(c.UserContext(), userID, watchID, uri); err != nil {
		return RespondWithError(c, err)
	}
	return c.JSON(fiber.Map{"message": "track deleted"})
}

func requestWatchIDs(c *fiber.Ctx) (int, int, error) {
	userID, err := requestUserID(c)
	if err != nil {
		return 0, 0, err
	}
	watchID, err := strconv.Atoi(c.Params("id"))
	if err != nil || watchID <= 0 {
		return 0, 0, domain.ErrBadRequest
	}
	return userID, watchID, nil
}

func trackQuery(c *fiber.Ctx) (int, int, bool, error) {
	offset := 0
	limit := domain.TrackPageSize
	removedOnly := false
	var err error
	if value := c.Query("offset"); value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			return 0, 0, false, domain.ErrBadRequest
		}
	}
	if value := c.Query("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > domain.TrackPageSize {
			return 0, 0, false, domain.ErrBadRequest
		}
	}
	if value := c.Query("removed"); value != "" {
		removedOnly, err = strconv.ParseBool(value)
		if err != nil {
			return 0, 0, false, domain.ErrBadRequest
		}
	}
	return offset, limit, removedOnly, nil
}
