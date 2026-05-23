package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

type authUsecase struct {
	authRepo       SessionRepository
	userRepo       UserRepository
	addressRepo    AddressRepository
	tokenProvider  TokenProvider
	passwordHasher PasswordHasher
}

func NewAuthUsecase(authRepo SessionRepository, userRepo UserRepository, addressRepo AddressRepository, tokenProvider TokenProvider, passwordHasher PasswordHasher) AuthUsecase {
	return &authUsecase{
		authRepo:       authRepo,
		userRepo:       userRepo,
		addressRepo:    addressRepo,
		tokenProvider:  tokenProvider,
		passwordHasher: passwordHasher,
	}
}

func (au *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	userID, _, err := au.tokenProvider.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("validateToken: parse: %w", err)
	}
	return au.userRepo.GetUserByID(ctx, userID)
}

func (au *authUsecase) generateTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	accessToken, err := au.tokenProvider.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return "", "", err
	}
	refreshToken, err := au.tokenProvider.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	session := &domain.UserSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30),
	}
	if err := au.authRepo.SaveSession(ctx, session); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (au *authUsecase) Register(ctx context.Context, phone, password, name string) (string, string, error) {
	if err := validatePhone(phone); err != nil {
		return "", "", err
	}
	if err := validatePassword(password); err != nil {
		return "", "", err
	}
	if err := validateName(name); err != nil {
		return "", "", err
	}

	existingUser, err := au.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		return "", "", err
	}
	if existingUser != nil {
		return "", "", ErrUserAlreadyExists
	}

	hashedPassword, err := au.passwordHasher.HashPassword(password)
	if err != nil {
		return "", "", err
	}

	user := &domain.User{
		Phone:        phone,
		Email:        "",
		PasswordHash: hashedPassword,
		Name:         name,
		Role:         domain.RoleUser,
	}

	if err := au.userRepo.CreateUser(ctx, user); err != nil {
		return "", "", err
	}

	return au.generateTokenPair(ctx, user)
}

func (au *authUsecase) Login(ctx context.Context, phone, password string) (string, string, error) {
	user, err := au.userRepo.GetUserByPhone(ctx, phone)

	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", ErrInvalidCredentials
	}
	if err := au.passwordHasher.CompareHashAndPassword(user.PasswordHash, password); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return au.generateTokenPair(ctx, user)
}

func (au *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	return au.authRepo.DeleteSession(ctx, refreshToken)
}

func (au *authUsecase) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := au.authRepo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	user, err := au.userRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return "", "", err
	}
	_ = au.authRepo.DeleteSession(ctx, refreshToken)
	return au.generateTokenPair(ctx, user)
}

func (au *authUsecase) UpdateProfile(ctx context.Context, params *domain.UpdateUserParams) (*domain.User, error) {
	return au.userRepo.UpdateUser(ctx, params)
}

func (au *authUsecase) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return au.userRepo.GetUserByID(ctx, id)
}

func (au *authUsecase) CreateAddress(ctx context.Context, addr *domain.Address) error {
	return au.addressRepo.CreateAddress(ctx, addr)
}

func (au *authUsecase) GetAddresses(ctx context.Context, userID int64) ([]domain.Address, error) {
	return au.addressRepo.GetAddressesByUserID(ctx, userID)
}

func (au *authUsecase) UpdateAddress(ctx context.Context, addr *domain.Address) error {
	return au.addressRepo.UpdateAddress(ctx, addr)
}

func (au *authUsecase) DeleteAddress(ctx context.Context, id int64) error {
	return au.addressRepo.DeleteAddress(ctx, id)
}
