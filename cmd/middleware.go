package cmd

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/improwised/datautil/pkg/auth"
	"github.com/improwised/datautil/pkg/models"
)

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := getTokenFromFiberRequest(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization token required",
			})
		}

		jwt := auth.NewJWT()
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}

func OptionalAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := getTokenFromFiberRequest(c)
		if token != "" {
			jwt := auth.NewJWT()
			claims, err := jwt.ValidateToken(token)
			if err == nil {
				c.Locals("user", claims)
			}
		}
		return c.Next()
	}
}

func getTokenFromFiberRequest(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}
	return c.Query("token")
}

func getUserFromContext(c *fiber.Ctx) *auth.Claims {
	if user, ok := c.Locals("user").(*auth.Claims); ok {
		return user
	}
	return nil
}

func checkPermission(c *fiber.Ctx, action string) bool {
	user := getUserFromContext(c)
	if user == nil {
		return false
	}
	tempUser := &models.User{Role: user.Role}
	return tempUser.HasPermission(action)
}
