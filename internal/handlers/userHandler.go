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

type deleteAddressReq struct {
	ID int64 `json:"id"`
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

// @Summary      Get Current User Profile
// @Description  Returns profile data of the authenticated user based on ID from token.
// @Tags         User
// @Security     BearerAuth
// @Produce      json
// @Success      200    {object}  domain.User   "User profile object"
// @Failure      401    {object}  ErrorResponse "Unauthorized"
// @Failure      404    {object}  ErrorResponse "User not found"
// @Failure      500    {object}  ErrorResponse "Internal server error"
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

// @Summary      Update User Profile
// @Description  Partially updates user profile fields (name, phone, email).
// @Tags         User
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input  body      domain.UpdateUserParams  true  "Profile Update Parameters"
// @Success      200    {object}  domain.User              "Updated user object"
// @Failure      400    {object}  ErrorResponse            "Validation error for provided parameters"
// @Failure      401    {object}  ErrorResponse            "Unauthorized"
// @Failure      404    {object}  ErrorResponse            "User not found"
// @Failure      500    {object}  ErrorResponse            "Failed to update profile"
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

// @Summary      Create Delivery Address
// @Description  Adds a new delivery address associated with the current user.
// @Tags         User
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input  body      domain.Address  true  "Address Data"
// @Success      201    {object}  domain.Address        "Successfully created address object"
// @Failure      400    {object}  ErrorResponse         "Invalid address validation parameters"
// @Failure      401    {object}  ErrorResponse         "Unauthorized"
// @Failure      500    {object}  ErrorResponse         "Failed to create address"
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

// @Summary      Get Delivery Addresses
// @Description  Retrieves all saved delivery addresses for the authenticated user.
// @Tags         User
// @Security     BearerAuth
// @Produce      json
// @Success      200    {array}   domain.Address  "Array of delivery addresses"
// @Failure      401    {object}  ErrorResponse   "Unauthorized"
// @Failure      500    {object}  ErrorResponse   "Internal server error"
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

// @Summary      Update Delivery Address
// @Description  Modifies fields of an existing delivery address.
// @Tags         User
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input  body      domain.UpdateAddressParams  true  "Address Data for Partial Update"
// @Success      200    "Address successfully updated"
// @Failure      400    {object}  ErrorResponse               "Invalid address validation parameters"
// @Failure      401    {object}  ErrorResponse               "Unauthorized"
// @Failure      404    {object}  ErrorResponse               "Address not found"
// @Failure      500    {object}  ErrorResponse               "Failed to update address"
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

// @Summary      Delete Delivery Address
// @Description  Removes a delivery address from the user's list by its ID.
// @Tags         User
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input  body      deleteAddressReq  true  "Target address ID to delete"
// @Success      204    "Address successfully deleted (No Content)"
// @Failure      400    {object}  ErrorResponse     "Invalid request body or non-positive ID"
// @Failure      401    {object}  ErrorResponse     "Unauthorized"
// @Failure      404    {object}  ErrorResponse     "Address not found"
// @Failure      500    {object}  ErrorResponse     "Failed to delete address"
// @Router       /api/user/address [delete]
func (h *userHandler) deleteAddress(c fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	var req deleteAddressReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	if req.ID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid address id"})
	}

	if err := h.userUC.DeleteAddress(c.Context(), userID, req.ID); err != nil {
		if errors.Is(err, usecase.ErrAddressNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "address not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to delete address"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
