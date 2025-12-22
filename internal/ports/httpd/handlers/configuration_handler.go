package handlers

import (
	"github.com/aritradeveops/porichoy/internal/core/service"
	"github.com/aritradeveops/porichoy/internal/pkg/translation"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) Configure(c *fiber.Ctx) error {
	var payload struct {
		RootUser RegisterUserPayload `json:"root_user"`
		RootApp  CreateAppPayload    `json:"root_app"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	err := h.service.Configure(c.Context(), service.ConfigurationPayload{
		RootUser: service.RegisterUserPayload(payload.RootUser),
		RootApp:  service.CreateAppPayload(payload.RootApp),
	})

	if err != nil {
		return err
	}

	return c.JSON(NewSuccessResponse(translation.Localize(c, "configuration.configure"), nil))
}
