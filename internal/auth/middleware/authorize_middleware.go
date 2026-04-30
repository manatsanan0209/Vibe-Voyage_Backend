package middleware

import (
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/token"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

const UserIDContextKey = "userID"

func Authorize() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if c.Method() == fiber.MethodGet && strings.HasPrefix(path, "/api/trip-suggestions") {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Next()
			}
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.Next()
			}

			tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if tokenStr == "" || tokenStr == "undefined" || tokenStr == "null" {
				return c.Next()
			}

			secret := os.Getenv("AUTH_TOKEN_SECRET")
			claims, err := token.Validate(tokenStr, secret)
			if err != nil {
				return unauthorized(c, err.Error())
			}
			if claims.UserID == 0 {
				return unauthorized(c, "invalid token claims")
			}

			c.Locals(UserIDContextKey, claims.UserID)
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return unauthorized(c, "missing or invalid authorization header")
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenStr == "" {
			return unauthorized(c, "missing or invalid authorization header")
		}

		secret := os.Getenv("AUTH_TOKEN_SECRET")
		claims, err := token.Validate(tokenStr, secret)
		if err != nil {
			return unauthorized(c, err.Error())
		}

		c.Locals(UserIDContextKey, claims.UserID)
		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) (uint, bool) {
	userID, ok := c.Locals(UserIDContextKey).(uint)
	if !ok || userID == 0 {
		return 0, false
	}

	return userID, true
}

func GetOptionalUserID(c *fiber.Ctx) (uint, bool, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return 0, false, nil
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, false, nil
	}

	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" || tokenStr == "undefined" || tokenStr == "null" {
		return 0, false, nil
	}

	secret := os.Getenv("AUTH_TOKEN_SECRET")
	claims, err := token.Validate(tokenStr, secret)
	if err != nil {
		return 0, false, err
	}
	if claims.UserID == 0 {
		return 0, false, errors.New("invalid token claims")
	}

	return claims.UserID, true, nil
}

func unauthorized(c *fiber.Ctx, errMsg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(dto.APIResponse[any]{
		Status:  fiber.StatusUnauthorized,
		Message: "unauthorized",
		Error:   errMsg,
	})
}
