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
)

type zapAdapter struct{}

func (za *zapAdapter) Error(msg string, fields ...any) { logger.Logger.Error(msg) }
func (za *zapAdapter) Warn(msg string, fields ...any)  { logger.Logger.Warn(msg) }

func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Logger.Warn("Файл .env не найден")
	}

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	app := fiber.New(fiber.Config{AppName: "MixFood Auth Service v1.0"})
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

	// 1. Инициализация инфраструктуры
	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, cfg.AccessTTL)
	passwordHasher := infrastructure.NewPasswordHasher()

	userRepo := infrastructure.NewUserRepository(conn, psql)
	sessionRepo := infrastructure.NewSessionRepository(conn, psql)
	addressRepo := infrastructure.NewAddressRepository(conn, psql)

	// 2. Инициализация Usecase (порядок аргументов должен совпадать с твоим usecase!)
	authUsecase := usecase.NewAuthUsecase(
		sessionRepo,
		userRepo,
		addressRepo,
		tokenProvider,
		passwordHasher,
	)

	// 3. Инициализация Handlers
	logAdapter := &zapAdapter{}
	authHandler := handlers.NewUsersHandler(authUsecase, tokenProvider, logAdapter)
	authHandler.RegisterRoutes(app)

	logger.Logger.Info(fmt.Sprintf("Сервер стартовал на :%s", cfg.ServerPort))
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Logger.Fatal("Сервер упал: " + err.Error())
	}
}
