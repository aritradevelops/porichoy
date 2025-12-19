package handlers

import (
	"github.com/aritradeveops/porichoy/internal/core/service"
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/aritradeveops/porichoy/internal/pkg/translation"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/authn"
	"github.com/gofiber/fiber/v2"
)

type CreateAppPayload struct {
	Name               string   `json:"name"`
	Domain             string   `json:"domain"`
	LandingUrl         string   `json:"landing_url"`
	Logo               string   `json:"logo"`
	RedirectUris       []string `json:"redirect_uris"`
	SuccessCallbackUrl string   `json:"success_callback_url"`
	ErrorCallbackUrl   string   `json:"error_callback_url"`
	JwtAlgo            string   `json:"jwt_algo"`
	JwtSecretResolver  string   `json:"jwt_secret_resolver"`
}

func (h *Handlers) CreateApp(c *fiber.Ctx) error {
	var payload CreateAppPayload
	err := c.BodyParser(&payload)
	if err != nil {
		return err
	}

	user, err := authn.GetUserFromContext(c)
	if err != nil {
		return err
	}

	app, err := h.service.CreateApp(c.Context(), user.UserID, service.CreateAppPayload(payload))
	if err != nil {
		logger.Error().Err(err)
		return err
	}

	return c.JSON(NewSuccessResponse(translation.Localize(c, "controller.create", map[string]string{
		"Resource": "App",
	}), app))
}
