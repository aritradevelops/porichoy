package handlers

import (
	"errors"

	"github.com/aritradeveops/porichoy/internal/core/service"
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/aritradeveops/porichoy/internal/pkg/translation"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/authn"
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

type Oauth2Payload struct {
	ClientID            string `query:"client_id"`
	ResponseType        string `query:"response_type"`
	RedirectURI         string `query:"redirect_uri"`
	CodeChallenge       string `query:"code_challenge"`
	CodeChallengeMethod string `query:"code_challenge_method"`
	State               string `query:"state"`
	LoginHint           string `query:"login_hint"`
	Nonce               string `query:"nonce"`
}

type Oauth2TokenPayload struct {
	ClientID     string `query:"client_id"`
	ClientSecret string `query:"client_secret"`
	GrantType    string `query:"grant_type"`
	Code         string `query:"code"`
	RedirectURI  string `query:"redirect_uri"`
}

func (h *Handlers) RegisterUser(c *fiber.Ctx) error {
	var payload RegisterUserPayload
	err := c.BodyParser(&payload)
	if err != nil {
		return err
	}
	user, err := h.service.RegisterUser(c.Context(), service.RegisterUserPayload(payload), false)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			return c.JSON(NewErrorResponse(translation.Localize(c, "user.exists"), err))
		}
		return err
	}
	return c.JSON(NewSuccessResponse(translation.Localize(c, "user.register"), user))
}

func (h *Handlers) LoginUser(c *fiber.Ctx) error {
	var payload LoginUserPayload
	err := c.BodyParser(&payload)
	if err != nil {
		return err
	}
	tokens, err := h.service.LoginUser(c.Context(), service.LoginUserPayload{
		Email:     payload.Email,
		Password:  payload.Password,
		UserAgent: c.Get("User-Agent"),
		UserIP:    c.IP(),
		Host:      c.Hostname(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidLoginCredentials) {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(NewErrorResponse(translation.Localize(c, "user.invalid_login_credentials"), err))
		} else if errors.Is(err, service.ErrDeactivatedUser) {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(NewErrorResponse(translation.Localize(c, "user.deactivated_user"), err))
		} else if errors.Is(err, service.ErrInvalidLoginMethod) {
			c.Status(fiber.StatusBadRequest)
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

func (h *Handlers) Oauth2(c *fiber.Ctx) error {
	var payload Oauth2Payload
	err := c.QueryParser(&payload)
	if err != nil {
		return err
	}
	user, err := authn.GetUserFromContext(c)
	if err != nil {
		return err
	}
	response, err := h.service.Oauth2(c.Context(), user.UserID, service.Oauth2Payload(payload))
	if err != nil {
		logger.Error().Err(err).Msg("oauth2 error")
		if response.OauthConfig.ErrorCallbackUrl != "" {
			return c.Redirect(response.OauthConfig.ErrorCallbackUrl)
		}
		// TODO: build an oauth error page and render error properly
		return err
	}

	if response.Oauth2CodeResponse != nil {
		return c.Redirect(response.Oauth2CodeResponse.RedirectURI)
	}

	return c.JSON(NewSuccessResponse(translation.Localize(c, "user.oauth2"), response))
}

func (h *Handlers) Token(c *fiber.Ctx) error {
	var payload Oauth2TokenPayload
	err := c.QueryParser(&payload)
	if err != nil {
		return err
	}
	tokens, err := h.service.Token(c.Context(), service.Oauth2TokenPayload{
		ClientID:     payload.ClientID,
		ClientSecret: payload.ClientSecret,
		GrantType:    payload.GrantType,
		Code:         payload.Code,
		RedirectURI:  payload.RedirectURI,
		UserAgent:    c.Get("User-Agent"),
		UserIP:       c.IP(),
	})
	if err != nil {
		return err
	}
	return c.JSON(NewSuccessResponse(translation.Localize(c, "user.token"), tokens))
}

func (h *Handlers) LogoutUser(c *fiber.Ctx) error {
	user, err := authn.GetUserFromContext(c)
	if err != nil {
		return err
	}
	err = h.service.LogoutUser(c.Context(), user.UserID)
	if err != nil {
		return err
	}
	c.ClearCookie("access_token")
	c.ClearCookie("refresh_token")
	return c.JSON(NewSuccessResponse(translation.Localize(c, "user.logout"), nil))
}
