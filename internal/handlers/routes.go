package handlers

import "github.com/gofiber/fiber/v3"

func (uh *usersHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")

	auth.Post("/register", fiber.Handler(uh.register))
	auth.Post("/login", fiber.Handler(uh.login))
	auth.Post("/logout", fiber.Handler(uh.logout))
	auth.Get("/refresh", fiber.Handler(uh.refresh))

	users := app.Group("/api/users", fiber.Handler(uh.AuthMiddleware()))
	users.Get("/me", fiber.Handler(uh.getMyProfile))
}
