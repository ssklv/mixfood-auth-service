package main

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-auth-service/internal/config"
	"github.com/ssklv/mixfood-auth-service/internal/handlers"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"

	_ "github.com/ssklv/mixfood-auth-service/docs"
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Error(fmt.Sprintf(msg+" %v", fields))
	}
}
func (za *zapAdapter) Warn(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Warn(fmt.Sprintf(msg+" %v", fields))
	}
}

// @title Mixfood Auth Service API
// @version 1.0
// @description API для аутентификации и работы с пользователями
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name jwt
func main() {
	logger.InitLogger()
	if logger.Logger != nil {
		defer logger.Logger.Sync()
	}

	if err := godotenv.Load(); err != nil {
		if logger.Logger != nil {
			logger.Logger.Warn("Файл .env не найден")
		}
	}

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	app := fiber.New(fiber.Config{
		AppName: "MixFood Auth Service v1.0",
	})
	//http://localhost:8080/swagger/
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Get("/swagger/", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(`<!DOCTYPE html>
            <html>
            <head><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css"></head>
            <body><div id="swagger-ui"></div>
            <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
            <script>
            SwaggerUIBundle({url: "/swagger/doc.json", dom_id: '#swagger-ui'});
            </script></body></html>`)
	})

	app.Get("/swagger/doc.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Ошибка БД: " + err.Error())
		}
	}
	defer conn.Close()

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, cfg.AccessTTL)
	passwordHasher := infrastructure.NewPasswordHasher()
	userRepo := infrastructure.NewUserRepository(conn, psql)
	sessionRepo := infrastructure.NewSessionRepository(conn, psql)
	addressRepo := infrastructure.NewAddressRepository(conn, psql)

	authUsecase := usecase.NewAuthUsecase(sessionRepo, userRepo, addressRepo, tokenProvider, passwordHasher)

	authHandler := handlers.NewUsersHandler(authUsecase, tokenProvider, &zapAdapter{})
	authHandler.RegisterRoutes(app)

	if logger.Logger != nil {
		logger.Logger.Info(fmt.Sprintf("Сервер стартовал на :%s", cfg.ServerPort))
	}

	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		if logger.Logger != nil {
			logger.Logger.Fatal("Сервер упал: " + err.Error())
		}
	}
}
