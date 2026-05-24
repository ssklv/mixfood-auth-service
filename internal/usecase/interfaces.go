package usecase

import (
	"context"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByPhone(ctx context.Context, phone string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateUser(ctx context.Context, params *domain.UpdateUserParams) (*domain.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

type AddressRepository interface {
	CreateAddress(ctx context.Context, addr *domain.Address) error
	GetAddressesByUserID(ctx context.Context, userID int64) ([]domain.Address, error)
	DeleteAddress(ctx context.Context, id int64) error
	UpdateAddress(ctx context.Context, addr *domain.Address) error
}

type SessionRepository interface {
	SaveSession(ctx context.Context, session *domain.UserSession) error
	GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error)
	DeleteSession(ctx context.Context, refreshToken string) error
}

// /
type AuthUsecase interface {
	Register(ctx context.Context, phone, password, name string) (string, string, error)
	Login(ctx context.Context, phone, password string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(ctx context.Context, tokenString string) (*domain.User, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)

	///для юзер репозитория
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateProfile(ctx context.Context, params *domain.UpdateUserParams) (*domain.User, error)

	//адрес
	CreateAddress(ctx context.Context, addr *domain.Address) error
	GetAddresses(ctx context.Context, userID int64) ([]domain.Address, error)
	UpdateAddress(ctx context.Context, addr *domain.Address) error
	DeleteAddress(ctx context.Context, id int64) error
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
