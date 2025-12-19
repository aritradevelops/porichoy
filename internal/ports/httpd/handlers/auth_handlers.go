package handlers

import (
	"errors"

	"github.com/aritradeveops/porichoy/internal/core/service"
	"github.com/aritradeveops/porichoy/internal/pkg/translation"
	"github.com/gofiber/fiber/v2"
)

type RegisterUserPayload struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type LoginUserPayload struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func (h *Handlers) RegisterUser(c *fiber.Ctx) error {
	var payload RegisterUserPayload
	err := c.BodyParser(&payload)
	if err != nil {
		return err
	}
	user, err := h.service.RegisterUser(c.Context(), service.RegisterUserPayload(payload))
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			return c.JSON(NewErrorResponse(translation.Localize(c, "user.exists"), err))
		}
		return err
	}
	return c.JSON(NewSuccessResponse(translation.Localize(c, "user.registered"), user))
}

func (h *Handlers) LoginUser(c *fiber.Ctx) error {
	var payload LoginUserPayload
	err := c.BodyParser(&payload)
	if err != nil {
		return err
	}
	tokens, err := h.service.LoginUser(c.Context(), service.LoginUserPayload(payload))
	if err != nil {
		if errors.Is(err, service.ErrInvalidLoginCredentials) {
			return c.JSON(NewErrorResponse(translation.Localize(c, "user.invalid_login_credentials"), err))
		} else if errors.Is(err, service.ErrDeactivatedUser) {
			return c.JSON(NewErrorResponse(translation.Localize(c, "user.deactivated_user"), err))
		} else if errors.Is(err, service.ErrInvalidLoginMethod) {
			return c.JSON(NewErrorResponse(translation.Localize(c, "user.invalid_login_method"), err))
		}
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		HTTPOnly: true,
		Expires:  tokens.AccessTokenExpiry,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HTTPOnly: true,
		Expires:  tokens.RefreshTokenExpiry,
	})

	return c.JSON(NewSuccessResponse(translation.Localize(c, "user.login"), tokens))
}

func (h *Handlers) GetUserProfile(c *fiber.Ctx) error {
	id := c.Locals("initiator").(string)

	return c.JSON(NewSuccessResponse(translation.Localize(c, "user.profile"), map[string]string{
		"id": id,
	}))
}
