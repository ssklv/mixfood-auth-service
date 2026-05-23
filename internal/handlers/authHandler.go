package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-auth-service/internal/domain"
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	accessToken, refreshToken, err := uh.usecase.Register(c.Context(), req.Phone, req.Password, req.Name)
	if err != nil {
		uh.log.Error("Registration error:", "err", err.Error())
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Этот номер телефона уже зарегистрирован"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	uh.setAuthCookies(c, accessToken, refreshToken)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"accessToken": accessToken,
	})
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
		uh.log.Error("login failed", err, "phone", req.Phone)
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверный номер телефона или пароль"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uh.setAuthCookies(c, accessToken, refreshToken)

	// Возвращаем токен для фронтенда
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"accessToken": accessToken,
	})
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

// @Summary Обновить профиль
// @Tags User
// @Accept json
// @Produce json
// @Param input body domain.UpdateUserParams true "Данные"
// @Success 200 {object} domain.User
// @Router /api/user/profile [patch]
func (uh *usersHandler) updateProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	var params domain.UpdateUserParams
	if err := c.Bind().Body(&params); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	params.ID = userID
	user, err := uh.usecase.UpdateProfile(c.Context(), &params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update profile"})
	}
	return c.JSON(user)
}

// @Summary Создать адрес
// @Tags User
// @Accept json
// @Produce json
// @Param input body domain.Address true "Данные адреса"
// @Success 201
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/user/address [post]
func (uh *usersHandler) createAddress(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	var addr domain.Address
	if err := c.Bind().Body(&addr); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	addr.UserID = userID
	if err := uh.usecase.CreateAddress(c.Context(), &addr); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create address"})
	}
	return c.SendStatus(201)
}

// @Summary Получить мои адреса
// @Tags User
// @Produce json
// @Success 200 {array} domain.Address
// @Failure 500 {object} ErrorResponse
// @Router /api/user/addresses [get]
func (uh *usersHandler) getMyAddresses(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	addresses, err := uh.usecase.GetAddresses(c.Context(), userID)
	if err != nil {
		uh.log.Error("failed to get addresses", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(addresses)
}

func (uh *usersHandler) getAccessToken(c fiber.Ctx) string {
	// 1. Проверяем заголовок Authorization (для твоего Axios-клиента)
	authHeader := c.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	// 2. Если заголовка нет, пробуем взять из куки (для старой логики)
	return c.Cookies(AccessCookie)
}

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
