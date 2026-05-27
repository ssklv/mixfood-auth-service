package main

//http://localhost:8080/swagger/
import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v3"
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
		logger.Logger.Sugar().Errorw(msg, fields...)
	}
}

func (za *zapAdapter) Warn(msg string, fields ...any) {
	if logger.Logger != nil {
		logger.Logger.Sugar().Warnw(msg, fields...)
	}
}

// @title                       Mixfood Auth Service API
// @version                     1.0
// @description                 API для аутентификации и работы с пользователями
// @host                        localhost:8080
// @BasePath                    /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Введите токен в формате: Bearer <token>

// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        access_token
// @description                 Токен доступа (access_token), автоматически извлекаемый из Cookie
func main() {
	logger.InitLogger()
	if logger.Logger != nil {
		defer logger.Logger.Sync()
	}

	if err := godotenv.Load(); err != nil && logger.Logger != nil {
		logger.Logger.Warn("Файл .env не найден")
	}

	cfg := config.Load()
	logAdapter := &zapAdapter{}

	conn, err := infrastructure.Connect(cfg.DatabaseURL)
	if err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Ошибка подключения к БД: " + err.Error())
	}
	defer conn.Close()

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	tokenProvider := infrastructure.NewTokenProvider(cfg.JWTSecret, cfg.AccessTTL)
	passwordHasher := infrastructure.NewPasswordHasher()

	userRepo := infrastructure.NewUserRepository(conn, psql)
	sessionRepo := infrastructure.NewSessionRepository(conn, psql)
	addressRepo := infrastructure.NewAddressRepository(conn, psql)

	authUsecase := usecase.NewAuthUsecase(sessionRepo, userRepo, tokenProvider, passwordHasher)
	userUsecase := usecase.NewUserUsecase(userRepo, addressRepo)

	app := fiber.New(fiber.Config{
		AppName: "MixFood Auth Service",
	})

	handlers.ConfigureApp(app, authUsecase, userUsecase, tokenProvider, logAdapter)

	if logger.Logger != nil {
		logger.Logger.Info(fmt.Sprintf("Сервер запущен на порту :%s", cfg.ServerPort))
	}

	if err := app.Listen(":" + cfg.ServerPort); err != nil && logger.Logger != nil {
		logger.Logger.Fatal("Ошибка сервера: " + err.Error())
	}
}
