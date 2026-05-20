package handlers

import "github.com/gofiber/fiber/v3"

func (uh *usersHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")
	auth.Post("/register", uh.register)
	auth.Post("/login", uh.login)
	auth.Post("/logout", uh.logout)
	auth.Get("/refresh", uh.refresh)

	users := app.Group("/api/users", uh.AuthMiddleware())
	users.Get("/me", uh.getMyProfile)
}
