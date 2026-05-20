package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
	"github.com/ssklv/mixfood-auth-service/internal/usecase"
	"github.com/ssklv/mixfood-auth-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthUsecase_Login_Success(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	mockSession := new(mocks.SessionRepository)
	mockHasher := new(mocks.PasswordHasher)
	mockToken := new(mocks.TokenProvider)

	uc := usecase.NewAuthUsecase(mockSession, mockUser, nil, mockToken, mockHasher)

	user := &domain.User{ID: 1, Role: domain.RoleUser, PasswordHash: "hash"}

	mockUser.On("GetUserByPhone", mock.Anything, "79991234567").Return(user, nil)
	mockHasher.On("CompareHashAndPassword", "hash", "pass").Return(nil)
	mockToken.On("GenerateAccessToken", int64(1), "user").Return("acc", nil)
	mockToken.On("GenerateRefreshToken").Return("ref", nil)
	mockSession.On("SaveSession", mock.Anything, mock.Anything).Return(nil)

	acc, ref, err := uc.Login(context.Background(), "79991234567", "pass")

	assert.NoError(t, err)
	assert.Equal(t, "acc", acc)
	assert.Equal(t, "ref", ref)
}

func TestAuthUsecase_Register_UserExists(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	uc := usecase.NewAuthUsecase(nil, mockUser, nil, nil, nil)

	mockUser.On("GetUserByPhone", mock.Anything, "123").Return(&domain.User{}, nil)

	_, _, err := uc.Register(context.Background(), "123", "pass", "Name")

	assert.Error(t, err)
}

func TestAuthUsecase_Logout(t *testing.T) {
	mockSession := new(mocks.SessionRepository)
	uc := usecase.NewAuthUsecase(mockSession, nil, nil, nil, nil)

	mockSession.On("DeleteSession", mock.Anything, "token").Return(nil)

	err := uc.Logout(context.Background(), "token")

	assert.NoError(t, err)
}

func TestAuthUsecase_ValidateToken_Error(t *testing.T) {
	mockToken := new(mocks.TokenProvider)
	uc := usecase.NewAuthUsecase(nil, nil, nil, mockToken, nil)

	mockToken.On("ParseToken", "bad_token").Return(int64(0), "", errors.New("invalid"))

	_, err := uc.ValidateToken(context.Background(), "bad_token")

	assert.Error(t, err)
}

func TestAuthUsecase_Login_InvalidPassword(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	mockHasher := new(mocks.PasswordHasher)
	uc := usecase.NewAuthUsecase(nil, mockUser, nil, nil, mockHasher)

	user := &domain.User{PasswordHash: "hash"}
	mockUser.On("GetUserByPhone", mock.Anything, "123").Return(user, nil)
	mockHasher.On("CompareHashAndPassword", "hash", "wrong").Return(errors.New("invalid"))

	_, _, err := uc.Login(context.Background(), "123", "wrong")
	assert.Error(t, err)
}

func TestAuthUsecase_AddressMethods(t *testing.T) {
	mockAddr := new(mocks.AddressRepository)
	uc := usecase.NewAuthUsecase(nil, nil, mockAddr, nil, nil)

	mockAddr.On("DeleteAddress", mock.Anything, int64(1)).Return(nil)
	err := uc.DeleteAddress(context.Background(), 1)
	assert.NoError(t, err)

	mockAddr.On("GetAddressesByUserID", mock.Anything, int64(1)).Return([]domain.Address{}, nil)
	addrs, err := uc.GetAddresses(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, addrs)
}

// Тест на ошибку при регистрации, если пользователь существует
func TestAuthUsecase_Register_Error(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	mockHasher := new(mocks.PasswordHasher)
	uc := usecase.NewAuthUsecase(nil, mockUser, nil, nil, mockHasher)

	// Ошибка: GetUserByPhone вернул пользователя (значит он уже есть)
	mockUser.On("GetUserByPhone", mock.Anything, "123").Return(&domain.User{}, nil)

	_, _, err := uc.Register(context.Background(), "123", "pass", "Name")
	assert.Error(t, err)
}

// Тест на ошибку при обновлении токенов (неверный сеанс)
func TestAuthUsecase_RefreshTokens_Error(t *testing.T) {
	mockSession := new(mocks.SessionRepository)
	uc := usecase.NewAuthUsecase(mockSession, nil, nil, nil, nil)

	mockSession.On("GetSessionByToken", mock.Anything, "bad_token").Return(nil, errors.New("not found"))

	_, _, err := uc.RefreshTokens(context.Background(), "bad_token")
	assert.Error(t, err)
}

// Тест на обновление профиля
func TestAuthUsecase_UpdateProfile(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	uc := usecase.NewAuthUsecase(nil, mockUser, nil, nil, nil)
	name := "newName"
	params := &domain.UpdateUserParams{ID: 1, Name: &name}
	mockUser.On("UpdateUser", mock.Anything, params).Return(&domain.User{Name: "NewName"}, nil)

	updated, err := uc.UpdateProfile(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, "NewName", updated.Name)
}

func TestAuthUsecase_AddressOperations(t *testing.T) {
	mockAddr := new(mocks.AddressRepository)
	uc := usecase.NewAuthUsecase(nil, nil, mockAddr, nil, nil)
	ctx := context.Background()

	// 1. Тест CreateAddress
	mockAddr.On("CreateAddress", ctx, mock.Anything).Return(nil)
	assert.NoError(t, uc.CreateAddress(ctx, &domain.Address{}))

	// 2. Тест UpdateAddress
	mockAddr.On("UpdateAddress", ctx, mock.Anything).Return(nil)
	assert.NoError(t, uc.UpdateAddress(ctx, &domain.Address{}))

	// 3. Тест DeleteAddress
	mockAddr.On("DeleteAddress", ctx, int64(1)).Return(nil)
	assert.NoError(t, uc.DeleteAddress(ctx, 1))
}

func TestAuthUsecase_RefreshTokens_Success(t *testing.T) {
	mockSession := new(mocks.SessionRepository)
	mockUser := new(mocks.UserRepository)
	mockToken := new(mocks.TokenProvider)

	uc := usecase.NewAuthUsecase(mockSession, mockUser, nil, mockToken, nil)
	ctx := context.Background()

	// Настройка успешного сценария
	mockSession.On("GetSessionByToken", ctx, "ref_token").Return(&domain.UserSession{UserID: 1}, nil)
	mockUser.On("GetUserByID", ctx, int64(1)).Return(&domain.User{ID: 1, Role: domain.RoleUser}, nil)
	mockSession.On("DeleteSession", ctx, "ref_token").Return(nil)

	// Ожидаем вызов generateTokenPair (внутренние вызовы токенов)
	mockToken.On("GenerateAccessToken", int64(1), "user").Return("new_acc", nil)
	mockToken.On("GenerateRefreshToken").Return("new_ref", nil)
	mockSession.On("SaveSession", ctx, mock.Anything).Return(nil)

	acc, ref, err := uc.RefreshTokens(ctx, "ref_token")

	assert.NoError(t, err)
	assert.Equal(t, "new_acc", acc)
	assert.Equal(t, "new_ref", ref)
}

func TestAuthUsecase_GenerateTokenPair_Error(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	mockSession := new(mocks.SessionRepository)
	mockHasher := new(mocks.PasswordHasher)
	mockToken := new(mocks.TokenProvider)

	uc := usecase.NewAuthUsecase(mockSession, mockUser, nil, mockToken, mockHasher)
	ctx := context.Background()

	// Настраиваем пользователя и пароль, чтобы Login прошел дальше
	user := &domain.User{ID: 1, Role: domain.RoleUser, PasswordHash: "hash"}
	mockUser.On("GetUserByPhone", mock.Anything, "123").Return(user, nil)
	mockHasher.On("CompareHashAndPassword", "hash", "pass").Return(nil)

	// А вот тут симулируем ошибку при генерации токена
	mockToken.On("GenerateAccessToken", int64(1), "user").Return("", errors.New("token err"))

	_, _, err := uc.Login(ctx, "123", "pass")

	assert.Error(t, err)
	assert.Equal(t, "token err", err.Error()) // Проверяем, что ошибка именно та
}

func TestAuthUsecase_ValidateToken_Success(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	mockToken := new(mocks.TokenProvider)
	uc := usecase.NewAuthUsecase(nil, mockUser, nil, mockToken, nil)
	ctx := context.Background()

	mockToken.On("ParseToken", "valid_token").Return(int64(1), "user", nil)
	mockUser.On("GetUserByID", ctx, int64(1)).Return(&domain.User{ID: 1}, nil)

	user, err := uc.ValidateToken(ctx, "valid_token")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestAuthUsecase_Login_SaveSessionError(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	mockSession := new(mocks.SessionRepository)
	mockHasher := new(mocks.PasswordHasher)
	mockToken := new(mocks.TokenProvider)

	uc := usecase.NewAuthUsecase(mockSession, mockUser, nil, mockToken, mockHasher)

	user := &domain.User{ID: 1, Role: domain.RoleUser, PasswordHash: "hash"}
	mockUser.On("GetUserByPhone", mock.Anything, "123").Return(user, nil)
	mockHasher.On("CompareHashAndPassword", "hash", "pass").Return(nil)
	mockToken.On("GenerateAccessToken", int64(1), "user").Return("acc", nil)
	mockToken.On("GenerateRefreshToken").Return("ref", nil)
	mockSession.On("SaveSession", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, _, err := uc.Login(context.Background(), "123", "pass")
	assert.Error(t, err)
}

func TestAuthUsecase_RefreshTokens_UserNotFound(t *testing.T) {
	mockSession := new(mocks.SessionRepository)
	mockUser := new(mocks.UserRepository)
	uc := usecase.NewAuthUsecase(mockSession, mockUser, nil, nil, nil)
	ctx := context.Background()

	// 1. Сессия найдена
	mockSession.On("GetSessionByToken", ctx, "ref").Return(&domain.UserSession{UserID: 1}, nil)
	// 2. А вот пользователя с таким ID нет
	mockUser.On("GetUserByID", ctx, int64(1)).Return(nil, errors.New("user not found"))

	_, _, err := uc.RefreshTokens(ctx, "ref")
	assert.Error(t, err)
}

func TestAuthUsecase_CreateAddress_Error(t *testing.T) {
	mockAddr := new(mocks.AddressRepository)
	uc := usecase.NewAuthUsecase(nil, nil, mockAddr, nil, nil)
	ctx := context.Background()

	mockAddr.On("CreateAddress", ctx, mock.Anything).Return(errors.New("db error"))

	err := uc.CreateAddress(ctx, &domain.Address{})
	assert.Error(t, err)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	mockUser := new(mocks.UserRepository)
	mockHasher := new(mocks.PasswordHasher)
	mockToken := new(mocks.TokenProvider)
	mockSession := new(mocks.SessionRepository)

	uc := usecase.NewAuthUsecase(mockSession, mockUser, nil, mockToken, mockHasher)
	ctx := context.Background()

	// Настраиваем: пользователь не найден, хэширование ок, сохранение ок
	mockUser.On("GetUserByPhone", ctx, "123").Return(nil, nil)
	mockHasher.On("HashPassword", "pass").Return("hashed", nil)
	mockUser.On("CreateUser", ctx, mock.Anything).Return(nil)

	// Моки для generateTokenPair
	mockToken.On("GenerateAccessToken", mock.Anything, mock.Anything).Return("acc", nil)
	mockToken.On("GenerateRefreshToken").Return("ref", nil)
	mockSession.On("SaveSession", ctx, mock.Anything).Return(nil)

	_, _, err := uc.Register(ctx, "123", "pass", "Name")
	assert.NoError(t, err)
}
