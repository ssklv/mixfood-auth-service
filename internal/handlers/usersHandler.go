package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
)

const (
	AccessCookie  = "access_token"
	RefreshCookie = "refresh_token"
)

type Logger interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

type UsersHandler interface {
	RegisterRoutes(app fiber.Router)
}

type usersHandler struct {
	usecase       usecase.AuthUsecase
	tokenProvider usecase.TokenProvider
	log           Logger
}

func NewUsersHandler(uc usecase.AuthUsecase, tp usecase.TokenProvider, log Logger) UsersHandler {
	return &usersHandler{usecase: uc, tokenProvider: tp, log: log}
}
