package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"retailpulse/apps/api/internal/domain"
	"retailpulse/apps/api/internal/platform/security"
)

const principalKey = "principal"

func Authenticate(tokens security.TokenManager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		header := ctx.Get(fiber.HeaderAuthorization)
		if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			return domain.ErrInvalidToken
		}
		principal, err := tokens.ParseAccessToken(strings.TrimSpace(header[7:]))
		if err != nil {
			return err
		}
		ctx.Locals(principalKey, principal)
		return ctx.Next()
	}
}

func Principal(ctx *fiber.Ctx) (domain.Principal, error) {
	value := ctx.Locals(principalKey)
	principal, ok := value.(domain.Principal)
	if !ok {
		return domain.Principal{}, domain.ErrInvalidToken
	}
	return principal, nil
}
