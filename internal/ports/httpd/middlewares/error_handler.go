package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/aritradeveops/porichoy/internal/core/validation"
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/aritradeveops/porichoy/internal/pkg/translation"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, e error) error {
		if e != nil {
			fmt.Println(e, reflect.TypeOf(e))
			logger.Error().Err(e)
			// handle fiber error
			if err, ok := e.(*fiber.Error); ok {
				c.Status(err.Code)
				return c.JSON(handlers.NewErrorResponse(
					translation.Localize(c, fmt.Sprintf("errors.%d", err.Code)),
					err,
				))
			}
			// handle validation error
			if errs, ok := e.(validation.ValidationErrors); ok {
				c.Status(http.StatusBadRequest)
				localizedErrors := fiber.Map{}
				for _, err := range errs {
					err.Message = translation.Localize(c, fmt.Sprintf("validation.%s", err.Code), map[string]any{
						"Param": err.Param,
						"Field": Labelize(err.Field),
					})
					localizedErrors[err.Field] = err
				}
				return c.JSON(handlers.NewErrorResponse(translation.Localize(c, fmt.Sprintf("errors.%d", http.StatusUnprocessableEntity)), localizedErrors))
			}

			// handle database error
			var pgErr *pgconn.PgError
			if errors.As(e, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					c.Status(http.StatusConflict)
					err := validation.ValidationError{
						Message: translation.Localize(c, "validation.unique", map[string]any{
							"Field": Labelize(getUniqueViolationField(pgErr.Detail)),
						}),
					}
					return c.JSON(handlers.NewErrorResponse(translation.Localize(c, fmt.Sprintf("errors.%d", http.StatusConflict)), fiber.Map{
						getUniqueViolationField(pgErr.Detail): err,
					}))
				}
			}
			c.Status(http.StatusInternalServerError)
			return c.JSON(handlers.NewErrorResponse(translation.Localize(c, fmt.Sprintf("errors.%d", http.StatusInternalServerError)), e))
		}
		return nil
	}
}

func Labelize(field string) string {
	formatted := []string{}
	for part := range strings.SplitSeq(field, "_") {
		formatted = append(formatted, strings.Title(part))
	}
	fmt.Println("field", field, "label", strings.Join(formatted, " "))
	return strings.Join(formatted, " ")
}

func getUniqueViolationField(detail string) string {
	// example: Key (email)=(test@gmail.com) already exists.
	re := regexp.MustCompile(`\(([^)]+)\)=\(([^)]+)\)`)
	matches := re.FindStringSubmatch(detail)

	if len(matches) == 3 {
		column := matches[1] // "email"
		// value := matches[2]  // "test@gmail.com"
		return column
	}
	return ""
}
