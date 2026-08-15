// Package rest is the REST adapter (CODING_STANDARDS.md §5): thin Fiber handlers that
// parse/validate a request, call exactly one bounded-context Service method, and serialize
// the result — no business logic lives here. This file holds the shared response envelope
// every route (success or failure) replies with, and the apperror.Error → HTTP mapping.
package rest

import (
	"errors"
	"net/http"

	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/gofiber/fiber/v2"
)

// envelope is the uniform response shape for every route (CODING_STANDARDS.md §5):
// Data is populated on success (Error is null), Error is populated on failure (Data is
// null) — never both.
type envelope struct {
	Data  any            `json:"data"`
	Error *envelopeError `json:"error"`
}

type envelopeError struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

// messages maps an apperror.Error's i18n key to English message text
// (CODING_STANDARDS.md §8). This is a placeholder, single-locale stand-in for the real
// translation-file mechanism §8 explicitly leaves undecided — good enough for every key
// this adapter currently raises, not a general i18n solution. A key with no entry here
// falls back to itself, so a response is never left with an empty message.
var messages = map[string]string{
	"tenant.not_found":                  "Tenant not found.",
	"tenant.domain_already_registered":  "This domain is already registered to a tenant.",
	"tenant.unresolved_domain":          "No tenant is registered for this domain.",
	"domain.not_found":                  "No tenant is registered for this domain.",
	"domain.forbidden":                  "You can only register a domain for your own tenant.",
	"identity.login_method_disabled":    "This tenant doesn't allow email/password sign-up.",
	"identity.email_already_registered": "An account with this email already exists.",
	"identity.system_app_not_found":     "This tenant isn't fully set up yet. Please contact your administrator.",
	"identity.password_too_long":        "Password must be 72 characters or fewer.",
	"identity.invalid_credentials":      "Incorrect email or password.",
	"identity.unauthenticated":          "You must be signed in to do that.",
	"authorization.forbidden":           "You don't have permission to do that.",
	"validation.invalid_body":           "The request body is malformed.",
	"validation.failed":                 "The request failed validation.",
	"internal.error":                    "Something went wrong. Please try again.",
}

func resolveMessage(key string) string {
	if msg, ok := messages[key]; ok {
		return msg
	}
	return key
}

// success writes a populated-data, null-error envelope.
func success(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(envelope{Data: data})
}

// fail writes a null-data, populated-error envelope. Known apperror.Errors carry their own
// status/key; anything else (a Postgres error, a context cancellation, ...) is treated as
// an opaque internal failure — its real message is never leaked to the client, only logged
// (once structured logging, zerolog per CODING_STANDARDS.md §1, is wired up — not yet).
func fail(c *fiber.Ctx, err error) error {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return c.Status(appErr.Status).JSON(envelope{
			Error: &envelopeError{Key: appErr.Key, Message: resolveMessage(appErr.Key)},
		})
	}
	return c.Status(http.StatusInternalServerError).JSON(envelope{
		Error: &envelopeError{Key: "internal.error", Message: resolveMessage("internal.error")},
	})
}
