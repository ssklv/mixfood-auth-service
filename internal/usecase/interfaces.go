package usecase

import (
	"context"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByPhone(ctx context.Context, phone string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)

	SaveSession(ctx context.Context, session *domain.UserSession) error
	GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error)
	DeleteSession(ctx context.Context, refreshToken string) error
}

type AuthUsecase interface {
	Register(ctx context.Context, phone, password, name string) (string, string, error)
	Login(ctx context.Context, phone, password string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(ctx context.Context, tokenString string) (*domain.User, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
}

type TokenProvider interface {
	GenerateAccessToken(userID int64, role string) (string, error)
	GenerateRefreshToken() (string, error)
	ParseToken(tokenString string) (int64, string, error)
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
}

type SessionRepository interface {
	SaveSession(ctx context.Context, session *domain.UserSession) error
	DeleteSession(ctx context.Context, refreshToken string) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*domain.User, error)
	UpdateUser(ctx context.Context, input *domain.UpdateUserParams) (*domain.User, error)
	DeleteUser(ctx context.Context, id int64) error
}
