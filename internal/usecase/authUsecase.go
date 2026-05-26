package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
)

type authUsecase struct {
	authRepo       SessionRepository
	userRepo       UserRepository
	tokenProvider  TokenProvider
	passwordHasher PasswordHasher
}

func NewAuthUsecase(authRepo SessionRepository, userRepo UserRepository, tokenProvider TokenProvider, passwordHasher PasswordHasher) AuthUsecase {
	return &authUsecase{
		authRepo:       authRepo,
		userRepo:       userRepo,
		tokenProvider:  tokenProvider,
		passwordHasher: passwordHasher,
	}
}

func (au *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	userID, _, err := au.tokenProvider.ParseToken(tokenString)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := au.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, ErrInternal
	}
	return user, nil
}

func (au *authUsecase) generateTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	accessToken, err := au.tokenProvider.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return "", "", ErrInternal
	}
	refreshToken, err := au.tokenProvider.GenerateRefreshToken()
	if err != nil {
		return "", "", ErrInternal
	}
	session := &domain.UserSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30),
	}
	if err := au.authRepo.SaveSession(ctx, session); err != nil {
		return "", "", ErrInternal
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

	_, err := au.userRepo.GetUserByPhone(ctx, phone)
	if err == nil {

		return "", "", ErrUserAlreadyExists
	}

	if !errors.Is(err, infrastructure.ErrUserNotFound) {
		return "", "", ErrInternal
	}

	hashedPassword, err := au.passwordHasher.HashPassword(password)
	if err != nil {
		return "", "", ErrInternal
	}

	user := &domain.User{
		Phone:        phone,
		Email:        "",
		PasswordHash: hashedPassword,
		Name:         name,
		Role:         domain.RoleUser,
	}

	if err := au.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, infrastructure.ErrDuplicatePhone) {
			return "", "", ErrUserAlreadyExists
		}
		return "", "", ErrInternal
	}

	return au.generateTokenPair(ctx, user)
}

func (au *authUsecase) Login(ctx context.Context, phone, password string) (string, string, error) {
	if err := validatePhone(phone); err != nil {
		return "", "", ErrInvalidCredentials
	}

	user, err := au.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", ErrInternal
	}

	if err := au.passwordHasher.CompareHashAndPassword(user.PasswordHash, password); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return au.generateTokenPair(ctx, user)
}

func (au *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	err := au.authRepo.DeleteSession(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, infrastructure.ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return ErrInternal
	}
	return nil
}

func (au *authUsecase) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := au.authRepo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, infrastructure.ErrSessionNotFound) {
			return "", "", ErrSessionNotFound
		}
		return "", "", ErrInternal
	}

	if time.Now().After(session.ExpiresAt) {
		_ = au.authRepo.DeleteSession(ctx, refreshToken)
		return "", "", ErrSessionExpired
	}

	user, err := au.userRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return "", "", ErrUserNotFound
		}
		return "", "", ErrInternal
	}

	_ = au.authRepo.DeleteSession(ctx, refreshToken)
	return au.generateTokenPair(ctx, user)
}
