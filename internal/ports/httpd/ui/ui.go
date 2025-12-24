package ui

import (
	"fmt"

	"github.com/aritradeveops/porichoy/internal/core/service"
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/authn"
	"github.com/gofiber/fiber/v2"
)

type UI struct {
	template string
	service  *service.Service
}
type OauthConsentPayload struct {
	ClientID string `query:"client_id" validate:"required"`
}

func New(template string, service *service.Service) *UI {
	return &UI{
		service:  service,
		template: template,
	}
}

func (u *UI) Index(c *fiber.Ctx) error {
	return c.Render("index", nil)
}

func (u *UI) Login(c *fiber.Ctx) error {
	return c.Render("login", nil)
}

func (u *UI) Register(c *fiber.Ctx) error {
	return c.Render("register", nil)
}

func (u *UI) Profile(c *fiber.Ctx) error {
	user, err := authn.GetUserFromContext(c)
	if err != nil {
		return err
	}
	fmt.Println("user", user)
	return c.Render("profile", user)
}

func (u *UI) OAuth2(c *fiber.Ctx) error {
	var payload OauthConsentPayload
	logger.Info().Any("queries", c.Queries()).Msg("oauth2")
	if err := c.QueryParser(&payload); err != nil {
		return err
	}
	logger.Info().Any("payload", payload).Msg("oauth2")
	resp, err := u.service.Oauth2ConsentResponse(c.Context(), service.Oauth2ConsentPayload{
		ClientID: payload.ClientID,
	})
	if err != nil {
		return err
	}
	return c.Render("oauth2", resp)
}
