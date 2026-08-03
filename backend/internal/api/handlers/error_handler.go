package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiarch/internal/shared/domain"
)

func RespondWithError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.HTTPCode).JSON(fiber.Map{"code": appErr.Code})
	}
	log.Error().Err(err).Msg("unexpected error")
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": domain.ErrInternal.Code})
}
