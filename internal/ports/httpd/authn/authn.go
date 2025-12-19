package authn

import (
	"fmt"
	"strings"

	"github.com/aritradeveops/porichoy/internal/config"
	"github.com/aritradeveops/porichoy/internal/core/jwtutil"
	"github.com/gofiber/fiber/v2"
)

const authUserKey = "auth_user"

func Middleware(conf config.JWT) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearer := c.Get("Authorization")
		if bearer == "" {
			return fiber.ErrUnauthorized
		}
		accessToken := strings.TrimPrefix(bearer, "Bearer ")
		payload, err := jwtutil.Verify(conf.Algorithm, accessToken, conf.VerifyingKeyResolver)
		if err != nil {
			return fiber.ErrUnauthorized
		}
		c.Locals(authUserKey, payload)
		return c.Next()
	}
}

func GetUserFromContext(c *fiber.Ctx) (jwtutil.JwtPayload, error) {
	userIn := c.Locals(authUserKey)
	if userIn == nil {
		return jwtutil.JwtPayload{}, fmt.Errorf("AuthenticatedUser is only available for protected routes")
	}
	payload, ok := userIn.(jwtutil.JwtPayload)
	if !ok {
		return jwtutil.JwtPayload{}, fmt.Errorf("AuthenticatedUser is only available for protected routes")
	}
	return payload, nil
}
