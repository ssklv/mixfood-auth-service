package main

import (
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/ssklv/mixfood-auth-service/internal/config"
	"github.com/ssklv/mixfood-auth-service/internal/handlers"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
	"github.com/ssklv/pizza-shared/pkg/logger"
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) { logger.Logger.Error(msg) }
func (za *zapAdapter) Warn(msg string, fields ...any)  { logger.Logger.Warn(msg) }

// @title MixFood Auth Service API
// @version 1.0
// @description Сервис аутентификации для доставки еды
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in cookie
// @name access_token
func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Logger.Warn("Файл .env не найден")
	}

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	app := fiber.New(fiber.Config{
		AppName: "MixFood Auth Service v1.0",
	})

	app.Get("/docs/*", adaptor.HTTPHandler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs")))))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
	}))

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Logger.Fatal("Ошибка БД: " + err.Error())
	}
	defer conn.Close()

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, cfg.AccessTTL)
	passwordHasher := infrastructure.NewPasswordHasher()

	userRepo := infrastructure.NewUserRepository(conn, psql)
	sessionRepo := infrastructure.NewSessionRepository(conn, psql)
	addressRepo := infrastructure.NewAddressRepository(conn, psql)

	authUsecase := usecase.NewAuthUsecase(
		sessionRepo,
		userRepo,
		addressRepo,
		tokenProvider,
		passwordHasher,
	)

	logAdapter := &zapAdapter{}
	authHandler := handlers.NewUsersHandler(authUsecase, tokenProvider, logAdapter)
	authHandler.RegisterRoutes(app)

	logger.Logger.Info(fmt.Sprintf("Сервер стартовал на :%s", cfg.ServerPort))
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Logger.Fatal("Сервер упал: " + err.Error())
	}
}
