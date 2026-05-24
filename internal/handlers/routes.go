package handlers

import (
	"github.com/gofiber/fiber/v3"
)

func (uh *usersHandler) RegisterRoutes(app fiber.Router) {
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/register", uh.register)
	auth.Post("/login", uh.login)
	auth.Get("/refresh", uh.refresh)

	protected := api.Group("/", uh.AuthMiddleware())
	protected.Post("/auth/logout", uh.logout)

	protected.Get("/users/me", uh.getMyProfile)
	protected.Patch("/user/profile", uh.updateProfile)
	protected.Post("/user/address", uh.createAddress)
	protected.Get("/user/addresses", uh.getMyAddresses)
}
