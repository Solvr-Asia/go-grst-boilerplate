package middleware

import (
	"strings"

	"go-grst-boilerplate/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

type AuthContext struct {
	UserID      string
	Email       string
	Roles       []string
	CompanyCode string
	Token       string
}

type AuthConfig struct {
	NeedAuth     bool
	AllowedRoles []string
}

type TokenValidator func(token string) (*AuthContext, error)

func AuthMiddleware(validator TokenValidator, config AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !config.NeedAuth {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return errors.Unauthorized("missing authorization header").FiberError(c)
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return errors.Unauthorized("invalid authorization format").FiberError(c)
		}

		token := parts[1]

		authCtx, err := validator(token)
		if err != nil {
			return errors.Unauthorized("invalid token").FiberError(c)
		}

		if len(config.AllowedRoles) > 0 {
			if !hasAnyRole(authCtx.Roles, config.AllowedRoles) {
				return errors.Forbidden("insufficient permissions").FiberError(c)
			}
		}

		c.Locals("auth", authCtx)

		return c.Next()
	}
}

func GetAuthContext(c *fiber.Ctx) (*AuthContext, bool) {
	auth, ok := c.Locals("auth").(*AuthContext)
	return auth, ok
}

func MustGetAuthContext(c *fiber.Ctx) *AuthContext {
	auth, ok := GetAuthContext(c)
	if !ok {
		panic("auth context not found")
	}
	return auth
}

func hasAnyRole(userRoles, allowedRoles []string) bool {
	roleSet := make(map[string]bool)
	for _, role := range userRoles {
		roleSet[role] = true
	}
	for _, allowed := range allowedRoles {
		if roleSet[allowed] {
			return true
		}
	}
	return false
}
