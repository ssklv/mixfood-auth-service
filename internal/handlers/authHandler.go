package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
)

type Logger interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

type UsersHandler interface {
	RegisterRoutes(app *fiber.App)
}

type usersHandler struct {
	usecase       usecase.AuthUsecase
	tokenProvider usecase.TokenProvider
	log           Logger
}

const (
	accessCookie  = "access_token"
	refreshCookie = "refresh_token"
)

func NewUsersHandler(uc usecase.AuthUsecase, tp usecase.TokenProvider, log Logger) UsersHandler {
	return &usersHandler{
		usecase:       uc,
		tokenProvider: tp,
		log:           log,
	}
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

func (uh *usersHandler) AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenStr := c.Cookies(accessCookie)
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Вы не авторизованы (токен отсутствует)"})
		}
		userID, role, err := uh.tokenProvider.ParseToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Сессия устарела или токен невалиден"})
		}

		c.Locals("userID", userID)
		c.Locals("userRole", role)

		return c.Next()
	}
}

func (uh *usersHandler) RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)
		if !ok || userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещен (роль не определена)"})
		}

		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return c.Next()
			}
		}

		uh.log.Warn("Попытка несанкционированного доступа", "user_id", c.Locals("userID"), "role", userRole)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "У вас недостаточно прав"})
	}
}

func (uh *usersHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")
	auth.Post("/register", uh.register)
	auth.Post("/login", uh.login)
	auth.Post("/logout", uh.logout)
	auth.Get("/refresh", uh.refresh)

	users := app.Group("/api/users")
	users.Get("/me", uh.AuthMiddleware(), uh.getMyProfile)

}

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

func (uh *usersHandler) logout(c fiber.Ctx) error {
	refreshToken := c.Cookies(refreshCookie)
	err := uh.usecase.Logout(c.Context(), refreshToken)

	uh.clearAuthCookies(c)

	if err != nil {
		uh.log.Error("logout failed", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to logout"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (uh *usersHandler) refresh(c fiber.Ctx) error {
	oldRefreshToken := c.Cookies(refreshCookie)
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

func (uh *usersHandler) getMyProfile(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Не удалось определить пользователя"})
	}

	user, err := uh.usecase.GetUserByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			uh.log.Error("user not found in getMyProfile", err, "userID", userID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		uh.log.Error("error getting profile", err, "userID", userID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (uh *usersHandler) setAuthCookies(c fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     accessCookie,
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 15),
		HTTPOnly: true,
		Secure:   false,
	})
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookie,
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HTTPOnly: true,
		Secure:   false,
	})
}

func (uh *usersHandler) clearAuthCookies(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     accessCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
	})
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
	})
}
