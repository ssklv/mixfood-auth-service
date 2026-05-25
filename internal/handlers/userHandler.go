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
	// Все роуты пользователей требуют обязательной авторизации
	user := router.Group("/user", authMiddleware)

	user.Get("/me", h.getMyProfile)
	user.Patch("/profile", h.updateProfile)
	user.Post("/address", h.createAddress)
	user.Get("/addresses", h.getMyAddresses)
	user.Put("/address", h.updateAddress)    // Добавили роут для обновления адреса
	user.Delete("/address", h.deleteAddress) // Добавили роут для удаления адреса
}

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

func (h *userHandler) getMyAddresses(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	addresses, err := h.userUC.GetAddresses(c.Context(), userID)
	if err != nil {
		h.log.Error("Failed to get addresses", "err", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "internal error"})
	}
	return c.JSON(addresses)
}

func (h *userHandler) updateAddress(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	var addr domain.Address
	if err := c.Bind().Body(&addr); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	if err := h.userUC.UpdateAddress(c.Context(), userID, &addr); err != nil {
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
