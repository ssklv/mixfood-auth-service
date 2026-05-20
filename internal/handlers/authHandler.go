package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
)

type registerReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}

// @Summary Регистрация
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body registerReq true "Данные пользователя"
// @Success 201
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/auth/register [post]
func (uh *usersHandler) register(c fiber.Ctx) error {
	var req registerReq
	if err := c.Bind().Body(&req); err != nil {
		uh.log.Error("invalid request body in register", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	accessToken, refreshToken, err := uh.usecase.Register(c.Context(), req.Phone, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPasswordTooWeak) ||
			errors.Is(err, usecase.ErrInvalidPhone) ||
			errors.Is(err, usecase.ErrInvalidName) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Этот номер телефона уже зарегистрирован"})
		}

		uh.log.Error("registration failed", err, "phone", req.Phone)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uh.setAuthCookies(c, accessToken, refreshToken)
	return c.SendStatus(fiber.StatusCreated)
}

// @Summary Логин
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body loginReq true "Телефон и пароль"
// @Success 200
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/login [post]
func (uh *usersHandler) login(c fiber.Ctx) error {
	var req loginReq
	if err := c.Bind().Body(&req); err != nil {
		uh.log.Error("invalid request body in login", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	accessToken, refreshToken, err := uh.usecase.Login(c.Context(), req.Phone, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверный номер телефона или пароль"})
		}

		uh.log.Error("login failed", err, "phone", req.Phone)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uh.setAuthCookies(c, accessToken, refreshToken)
	return c.SendStatus(fiber.StatusOK)
}

// @Summary Выход
// @Tags Auth
// @Success 204
// @Router /api/auth/logout [post]
func (uh *usersHandler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies(RefreshCookie)
	err := uh.usecase.Logout(c.Context(), refreshToken)

	uh.clearAuthCookies(c)

	if err != nil {
		uh.log.Error("logout failed", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to logout"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary Обновление токенов
// @Tags Auth
// @Success 200
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/refresh [get]
func (uh *usersHandler) refresh(c fiber.Ctx) error {
	oldRefreshToken := c.Cookies(RefreshCookie)
	if oldRefreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "empty refresh token"})
	}

	accessToken, newRefreshToken, err := uh.usecase.RefreshTokens(c.Context(), oldRefreshToken)
	if err != nil {
		uh.log.Warn("token refresh failed", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "session expired or invalid"})
	}

	uh.setAuthCookies(c, accessToken, newRefreshToken)
	return c.SendStatus(fiber.StatusOK)
}

// Утилитные методы
func (uh *usersHandler) setAuthCookies(c fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     AccessCookie,
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 15),
		HTTPOnly: true,
		Secure:   false,
	})
	c.Cookie(&fiber.Cookie{
		Name:     RefreshCookie,
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HTTPOnly: true,
		Secure:   false,
	})
}

func (uh *usersHandler) clearAuthCookies(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     AccessCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
	})
	c.Cookie(&fiber.Cookie{
		Name:     RefreshCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
	})
}
