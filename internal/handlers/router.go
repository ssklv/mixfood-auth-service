package handlers

import (
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
)

func ConfigureApp(
	app *fiber.App,
	authUC usecase.AuthUsecase,
	userUC usecase.UserUsecase,
	tokenProvider usecase.TokenProvider,
	log Logger,
) {
	app.Use(recover.New())
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:5173", "http://localhost:8082"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Get("/swagger/*", swaggo.HandlerDefault)

	authMiddleware := NewAuthMiddleware(tokenProvider, log)

	apiGroup := app.Group("/api")

	NewAuthHandler(authUC, tokenProvider, log).RegisterRoutes(apiGroup, authMiddleware)
	NewUserHandler(userUC, log).RegisterRoutes(apiGroup, authMiddleware)
}
