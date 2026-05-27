package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-auth-service/internal/domain"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
)

type userHandler struct {
	userUC usecase.UserUsecase
	log    Logger
}

func NewUserHandler(userUC usecase.UserUsecase, log Logger) *userHandler {
	return &userHandler{
		userUC: userUC,
		log:    log,
	}
}

func (h *userHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	user := router.Group("/user", authMiddleware)

	user.Get("/me", h.getMyProfile)
	user.Patch("/profile", h.updateProfile)
	user.Post("/address", h.createAddress)
	user.Get("/addresses", h.getMyAddresses)
	user.Put("/address", h.updateAddress)
	user.Delete("/address", h.deleteAddress)
}

// @Summary      Профиль текущего пользователя
// @Description  Возвращает данные профиля авторизованного пользователя на основе ID из токена.
// @Tags         User
// @Security     BearerAuth
// @Security     CookieAuth
// @Produce      json
// @Success      200    {object}  domain.User   "Объект пользователя"
// @Failure      401    {object}  ErrorResponse "Unauthorized"
// @Failure      404    {object}  ErrorResponse "user not found"
// @Failure      500    {object}  ErrorResponse "internal error"
// @Router       /api/user/me [get]
func (h *userHandler) getMyProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	user, err := h.userUC.GetProfile(c.Context(), userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "internal error"})
	}
	return c.JSON(user)
}

// @Summary      Обновить профиль
// @Description  Частичное обновление полей профиля пользователя (имя, телефон, email).
// @Tags         User
// @Security     BearerAuth
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        input  body      domain.UpdateUserParams  true  "Данные профиля"
// @Success      200    {object}  domain.User              "Обновленный объект пользователя"
// @Failure      400    {object}  ErrorResponse            "Ошибка валидации переданных параметров"
// @Failure      401    {object}  ErrorResponse            "Unauthorized"
// @Failure      404    {object}  ErrorResponse            "user not found"
// @Failure      500    {object}  ErrorResponse            "failed to update profile"
// @Router       /api/user/profile [patch]
func (h *userHandler) updateProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	var params domain.UpdateUserParams
	if err := c.Bind().Body(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	params.ID = userID
	user, err := h.userUC.UpdateProfile(c.Context(), &params)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidPhone) || errors.Is(err, usecase.ErrInvalidName) || errors.Is(err, usecase.ErrInvalidEmail) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to update profile"})
	}
	return c.JSON(user)
}

// @Summary      Создать адрес
// @Description  Добавление нового адреса доставки, связанного с текущим пользователем.
// @Tags         User
// @Security     BearerAuth
// @Security     CookieAuth
// @Accept       json
// @Produce      json
// @Param        input  body      domain.Address  true  "Данные адреса"
// @Success      201    {object}  domain.Address        "Успешно созданный адрес"
// @Failure      400    {object}  ErrorResponse         "Невалидный адрес"
// @Failure      401    {object}  ErrorResponse         "Unauthorized"
// @Failure      500    {object}  ErrorResponse         "failed to create address"
// @Router       /api/user/address [post]
func (h *userHandler) createAddress(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	var addr domain.Address
	if err := c.Bind().Body(&addr); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	addr.UserID = userID
	if err := h.userUC.CreateAddress(c.Context(), &addr); err != nil {
		if errors.Is(err, usecase.ErrInvalidAddress) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to create address"})
	}
	return c.Status(fiber.StatusCreated).JSON(addr)
}

// @Summary      Список адресов
// @Description  Получение всех сохраненных адресов доставки авторизованного пользователя.
// @Tags         User
// @Security     BearerAuth
// @Security     CookieAuth
// @Produce      json
// @Success      200    {array}   domain.Address "Массив адресов"
// @Failure      401    {object}  ErrorResponse  "Unauthorized"
// @Failure      500    {object}  ErrorResponse  "internal error"
// @Router       /api/user/addresses [get]
func (h *userHandler) getMyAddresses(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	addresses, err := h.userUC.GetAddresses(c.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get addresses", "err", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "internal error"})
	}
	return c.JSON(addresses)
}

// @Summary      Обновить адрес
// @Description  Модификация полей существующего адреса доставки.
// @Tags         User
// @Security     BearerAuth
// @Security     CookieAuth
// @Accept       json
// @Param        input  body      domain.Address  true  "Данные адреса (включая ID адреса)"
// @Success      200    "Адрес успешно обновлен"
// @Failure      400    {object}  ErrorResponse   "Невалидный адрес"
// @Failure      401    {object}  ErrorResponse   "Unauthorized"
// @Failure      404    {object}  ErrorResponse   "address not found"
// @Failure      500    {object}  ErrorResponse   "failed to update address"
// @Router       /api/user/address [put]
func (h *userHandler) updateAddress(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	var params domain.UpdateAddressParams
	if err := c.Bind().Body(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	if err := h.userUC.UpdateAddress(c.Context(), userID, &params); err != nil {
		if errors.Is(err, usecase.ErrInvalidAddress) {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		}
		if errors.Is(err, usecase.ErrAddressNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "address not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to update address"})
	}
	return c.SendStatus(fiber.StatusOK)
}

type deleteAddressReq struct {
	ID int64 `json:"id"`
}

// @Summary      Удалить адрес
// @Description  Удаление адреса доставки из списка пользователя по ID.
// @Tags         User
// @Security     BearerAuth
// @Security     CookieAuth
// @Accept       json
// @Param        input  body      deleteAddressReq  true  "ID адреса для удаления"
// @Success      204    "Адрес успешно удален (No Content)"
// @Failure      400    {object}  ErrorResponse "invalid request body"
// @Failure      401    {object}  ErrorResponse "Unauthorized"
// @Failure      404    {object}  ErrorResponse "address not found"
// @Failure      500    {object}  ErrorResponse "failed to delete address"
// @Router       /api/user/address [delete]
func (h *userHandler) deleteAddress(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	var req deleteAddressReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	if err := h.userUC.DeleteAddress(c.Context(), userID, req.ID); err != nil {
		if errors.Is(err, usecase.ErrAddressNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "address not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to delete address"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
