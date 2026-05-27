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
	GetAddressByID(ctx context.Context, id int64) (*domain.Address, error)
	UpdateAddress(ctx context.Context, params *domain.UpdateAddressParams) error
	DeleteAddress(ctx context.Context, id int64) error
}

type SessionRepository interface {
	SaveSession(ctx context.Context, session *domain.UserSession) error
	GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error)
	DeleteSession(ctx context.Context, refreshToken string) error
}

type AuthUsecase interface {
	Register(ctx context.Context, phone, password, name string) (accessToken string, refreshToken string, err error)
	Login(ctx context.Context, phone, password string) (accessToken string, refreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshTokens(ctx context.Context, tokenInput string) (accessToken string, refreshToken string, err error)
	ValidateToken(ctx context.Context, tokenString string) (*domain.User, error)
}

type UserUsecase interface {
	GetProfile(ctx context.Context, id int64) (*domain.User, error)
	UpdateProfile(ctx context.Context, params *domain.UpdateUserParams) (*domain.User, error)
	CreateAddress(ctx context.Context, addr *domain.Address) error
	GetAddresses(ctx context.Context, userID int64) ([]domain.Address, error)
	UpdateAddress(ctx context.Context, userID int64, params *domain.UpdateAddressParams) error
	DeleteAddress(ctx context.Context, userID int64, addressID int64) error
}

type TokenProvider interface {
	GenerateAccessToken(userID int64, role string) (string, error)
	GenerateRefreshToken() (string, error)
	ParseToken(tokenString string) (userID int64, role string, err error)
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash, password string) error
}
