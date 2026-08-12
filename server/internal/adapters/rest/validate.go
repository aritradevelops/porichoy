package rest

import (
	"net/http"

	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// validate is a single shared validator instance (CODING_STANDARDS.md §1 —
// go-playground/validator for request validation) — safe for concurrent use across every
// handler in this package, per the library's own documentation.
var validate = validator.New()

// bindAndValidate decodes c's request body into dst and runs struct validation tags on it,
// collapsing both failure modes (malformed JSON, failed validation) into the two
// apperror.Error keys every handler needs to hand to fail(): a caller-facing detail isn't
// worth exposing beyond that (CODING_STANDARDS.md §5 keeps error messages i18n-keyed, not
// raw library/parser output).
func bindAndValidate(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return apperror.New("validation.invalid_body", http.StatusBadRequest)
	}
	if err := validate.Struct(dst); err != nil {
		return apperror.New("validation.failed", http.StatusBadRequest)
	}
	return nil
}
