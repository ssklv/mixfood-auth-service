package handlers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
	_ "github.com/ssklv/mixfood-auth-service/internal/domain"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
)

func (uh *usersHandler) AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else {
			token = c.Cookies(AccessCookie)
		}

		fmt.Printf("DEBUG: Middleware получил токен: '%s'\n", token)

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		userID, role, err := uh.tokenProvider.ParseToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		c.Locals("userID", userID)
		c.Locals("userRole", role)

		return c.Next()
	}
}

// @Summary Получить профиль пользователя
// @Tags Profile
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.User
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /api/users/me [get]
func (uh *usersHandler) getMyProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	user, err := uh.usecase.GetUserByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(user)
}
