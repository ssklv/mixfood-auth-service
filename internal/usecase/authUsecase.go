package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/ssklv/food-delivery-backend/internal/domain"
)

//ошибки добавить

type authUsecase struct {
	repository     AuthRepository
	tokenProvider  TokenProvider
	passwordHasher PasswordHasher
}

func NewAuthUsecase(rep AuthRepository, tokenProvider TokenProvider, passwordHasher PasswordHasher) AuthUsecase {
	return &authUsecase{
		repository:     rep,
		tokenProvider:  tokenProvider,
		passwordHasher: passwordHasher,
	}
}

func (au *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	userID, err := au.tokenProvider.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("validateToken: parse: %w", err)
	}

	user, err := au.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("validateToken: get user: %w", err)
	}

	return user, nil
}

func (au *authUsecase) generateTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	accessToken, err := au.tokenProvider.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := au.tokenProvider.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	session := &domain.UserSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30),
	}

	if err := au.repository.SaveSession(ctx, session); err != nil {
		return "", "", fmt.Errorf("save session: %w", err)
	}

	return accessToken, refreshToken, nil
}

// //
func (au *authUsecase) Register(ctx context.Context, phone, password, name string) (string, string, error) {
	if err := validatePassword(password); err != nil {
		return "", "", err
	}
	if err := validatePhone(phone); err != nil {
		return "", "", err
	}
	if err := validateName(name); err != nil {
		return "", "", err
	}

	hashedPassword, err := au.passwordHasher.HashPassword(password)
	if err != nil {
		return "", "", fmt.Errorf("register: hash password: %w", err)
	}

	user := &domain.User{
		Phone:        phone,
		PasswordHash: hashedPassword,
		Name:         name,
		Role:         domain.RoleUser, ////
	}

	err = au.repository.CreateUser(ctx, user)
	if err != nil {
		return "", "", fmt.Errorf("register: create user: %w", err)
	}

	accessToken, refreshToken, err := au.generateTokenPair(ctx, user)
	if err != nil {
		return "", "", fmt.Errorf("generate tokens: %w", err)
	}

	return accessToken, refreshToken, nil
}

// //////
func (au *authUsecase) Login(ctx context.Context, phone, password string) (string, string, error) {
	user, err := au.repository.GetUserByPhone(ctx, phone)
	if err != nil {
		return "", "", fmt.Errorf("hash password: %w", err) ////ПЕРЕПИСАТЬ
	}

	err = au.passwordHasher.CompareHashAndPassword(user.PasswordHash, password)
	if err != nil {
		return "", "", fmt.Errorf("hash password: %w", err) ///ПЕРЕПИСАТЬ
	}

	accessToken, refreshToken, err := au.generateTokenPair(ctx, user)
	if err != nil {
		return "", "", fmt.Errorf("login: generate tokens: %w", err)
	}
	return accessToken, refreshToken, nil
}

// //////
func (au *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	err := au.repository.DeleteSession(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}

func (au *authUsecase) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := au.repository.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("refresh: session not found: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = au.repository.DeleteSession(ctx, refreshToken) // Удаляем протухшую
		return "", "", fmt.Errorf("refresh: session expired")
	}
	user, err := au.repository.GetUserByID(ctx, session.UserID)
	if err != nil {
		return "", "", fmt.Errorf("refresh: get user: %w", err)
	}

	_ = au.repository.DeleteSession(ctx, refreshToken)

	return au.generateTokenPair(ctx, user)
}
