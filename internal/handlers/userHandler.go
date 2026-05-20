package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
)

func (uh *usersHandler) AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenStr := c.Cookies("access_token")
		userID, role, err := uh.tokenProvider.ParseToken(tokenStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
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
// @Success 200 {object} domain.User "Данные профиля"
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
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
