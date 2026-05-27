package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
)

type authHandler struct {
	authUC        usecase.AuthUsecase
	tokenProvider usecase.TokenProvider
	log           Logger
}

type registerReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func NewAuthHandler(authUC usecase.AuthUsecase, tp usecase.TokenProvider, log Logger) *authHandler {
	return &authHandler{
		authUC:        authUC,
		tokenProvider: tp,
		log:           log,
	}
}

func (h *authHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	auth := router.Group("/auth")

	auth.Post("/register", h.register)
	auth.Post("/login", h.login)
	auth.Get("/refresh", h.refresh)

	auth.Post("/logout", authMiddleware, h.logout)
}

// @Summary      Регистрация пользователя
// @Description  Создание нового аккаунта, генерация и автоматическая установка JWT-токенов в куки.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input  body      registerReq  true  "Данные для регистрации"
// @Success      201    {object}  map[string]string   "Возвращает accessToken в формате JSON"
// @Failure      400    {object}  ErrorResponse       "Невалидный формат запроса или слабый пароль"
// @Failure      409    {object}  ErrorResponse       "Этот номер телефона уже зарегистрирован"
// @Failure      500    {object}  ErrorResponse       "Внутренняя ошибка сервера"
// @Router       /api/auth/register [post]
func (h *authHandler) register(c fiber.Ctx) error {
	var req registerReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	accessToken, refreshToken, err := h.authUC.Register(c.Context(), req.Phone, req.Password, req.Name)
	if err != nil {
		h.log.Error("Registration error", "err", err.Error())
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: "Этот номер телефона уже зарегистрирован"})
		}
		if errors.Is(err, usecase.ErrInvalidPhone) || errors.Is(err, usecase.ErrInvalidName) || errors.Is(err, usecase.ErrInvalidPasswordTooWeak) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "internal error"})
	}

	h.setAuthCookies(c, accessToken, refreshToken)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"accessToken": accessToken,
	})
}

// @Summary      Авторизация (Вход)
// @Description  Аутентификация по номеру телефона и паролю. Устанавливает токены в httpOnly куки.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input  body      loginReq  true  "Учетные данные"
// @Success      200    {object}  map[string]string "Возвращает accessToken в формате JSON"
// @Failure      400    {object}  ErrorResponse     "invalid request body"
// @Failure      401    {object}  ErrorResponse     "Неверный номер телефона или пароль"
// @Failure      500    {object}  ErrorResponse     "internal server error"
// @Router       /api/auth/login [post]
func (h *authHandler) login(c fiber.Ctx) error {
	var req loginReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	accessToken, refreshToken, err := h.authUC.Login(c.Context(), req.Phone, req.Password)
	if err != nil {
		h.log.Warn("Login failed", "phone", req.Phone, "err", err.Error())
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "Неверный номер телефона или пароль"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "internal server error"})
	}

	h.setAuthCookies(c, accessToken, refreshToken)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"accessToken": accessToken,
	})
}

// @Summary      Выход из системы
// @Description  Удаляет сессию из базы данных и инвалидирует авторизационные куки на клиенте.
// @Tags         Auth
// @Security     BearerAuth
// @Security     CookieAuth
// @Success      204    "Успешный выход без тела ответа"
// @Failure      404    {object}  ErrorResponse "active session not found"
// @Failure      500    {object}  ErrorResponse "failed to logout"
// @Router       /api/auth/logout [post]
func (h *authHandler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies(RefreshCookie)
	err := h.authUC.Logout(c.Context(), refreshToken)

	h.clearAuthCookies(c)

	if err != nil {
		h.log.Error("Logout failed", "err", err.Error())
		if errors.Is(err, usecase.ErrSessionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "active session not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to logout"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary      Обновление сессии (Refresh)
// @Description  Принимает refresh_token из куки, проверяет его валидность и генерирует новую пару токенов.
// @Tags         Auth
// @Success      200    "Токены успешно обновлены"
// @Failure      401    {object}  ErrorResponse "session expired / invalid"
// @Router       /api/auth/refresh [get]
func (h *authHandler) refresh(c fiber.Ctx) error {
	oldRefreshToken := c.Cookies(RefreshCookie)
	if oldRefreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "empty refresh token"})
	}

	accessToken, newRefreshToken, err := h.authUC.RefreshTokens(c.Context(), oldRefreshToken)
	if err != nil {
		h.log.Warn("Token refresh failed", "err", err.Error())
		if errors.Is(err, usecase.ErrSessionExpired) {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "session expired"})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "session not found or invalid"})
	}

	h.setAuthCookies(c, accessToken, newRefreshToken)
	return c.SendStatus(fiber.StatusOK)
}

func (h *authHandler) setAuthCookies(c fiber.Ctx, accessToken, refreshToken string) {
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

func (h *authHandler) clearAuthCookies(c fiber.Ctx) {
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
