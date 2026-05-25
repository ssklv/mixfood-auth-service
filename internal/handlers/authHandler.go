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

type authHandler struct {
	authUC        usecase.AuthUsecase
	tokenProvider usecase.TokenProvider
	log           Logger
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

	// Логаут требует, чтобы пользователь был авторизован
	auth.Post("/logout", authMiddleware, h.logout)
}

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

func (h *authHandler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies(RefreshCookie)
	err := h.authUC.Logout(c.Context(), refreshToken)

	h.clearAuthCookies(c)

	if err != nil {
		h.log.Error("Logout failed", "err", err.Error())
		// Мы всё равно очистили куки клиента, но возвращаем ошибку, если сессии не было в БД
		if errors.Is(err, usecase.ErrSessionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "active session not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to logout"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

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
