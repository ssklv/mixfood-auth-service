package usecase

//go test -v ./internal/usecase/
//go test -cover ./internal/usecase/
import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
	"github.com/ssklv/mixfood-auth-service/internal/usecase/mocks"
)

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := mocks.NewUserRepository(t)
	mockSessionRepo := mocks.NewSessionRepository(t)
	mockTokenProvider := mocks.NewTokenProvider(t)
	mockHasher := mocks.NewPasswordHasher(t)

	mockUserRepo.On("GetUserByPhone", ctx, "89991234567").Return(nil, infrastructure.ErrUserNotFound)

	mockHasher.On("HashPassword", "password123").Return("secret_hash", nil)

	expectedUser := &domain.User{
		Phone:        "89991234567",
		PasswordHash: "secret_hash",
		Name:         "Илья",
		Role:         domain.RoleUser,
	}
	mockUserRepo.On("CreateUser", ctx, expectedUser).Return(nil)

	mockTokenProvider.On("GenerateAccessToken", int64(0), "user").Return("access-jwt", nil)
	mockTokenProvider.On("GenerateRefreshToken").Return("refresh-jwt", nil)

	mockSessionRepo.On("SaveSession", ctx, mock.MatchedBy(func(s *domain.UserSession) bool {
		return s.RefreshToken == "refresh-jwt"
	})).Return(nil)

	uc := NewAuthUsecase(mockSessionRepo, mockUserRepo, mockTokenProvider, mockHasher)

	accessToken, refreshToken, err := uc.Register(ctx, "89991234567", "password123", "Илья")

	assert.NoError(t, err)
	assert.Equal(t, "access-jwt", accessToken)
	assert.Equal(t, "refresh-jwt", refreshToken)
}

func TestRegister_DuplicatePhone(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := mocks.NewUserRepository(t)
	mockSessionRepo := mocks.NewSessionRepository(t)
	mockTokenProvider := mocks.NewTokenProvider(t)
	mockHasher := mocks.NewPasswordHasher(t)

	existingUser := &domain.User{ID: 1, Phone: "89991234567"}
	mockUserRepo.On("GetUserByPhone", ctx, "89991234567").Return(existingUser, nil)

	uc := NewAuthUsecase(mockSessionRepo, mockUserRepo, mockTokenProvider, mockHasher)

	accessToken, refreshToken, err := uc.Register(ctx, "89991234567", "password123", "Илья")

	assert.ErrorIs(t, err, ErrUserAlreadyExists)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

// на логин
func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := mocks.NewUserRepository(t)
	mockSessionRepo := mocks.NewSessionRepository(t)
	mockTokenProvider := mocks.NewTokenProvider(t)
	mockHasher := mocks.NewPasswordHasher(t)

	dbUser := &domain.User{
		ID:           10,
		Phone:        "89991234567",
		PasswordHash: "secret_hash",
		Role:         domain.RoleUser,
	}
	mockUserRepo.On("GetUserByPhone", ctx, "89991234567").Return(dbUser, nil)

	mockHasher.On("CompareHashAndPassword", "secret_hash", "password123").Return(nil)

	mockTokenProvider.On("GenerateAccessToken", int64(10), "user").Return("access-jwt", nil)
	mockTokenProvider.On("GenerateRefreshToken").Return("refresh-jwt", nil)
	mockSessionRepo.On("SaveSession", ctx, mock.Anything).Return(nil)

	uc := NewAuthUsecase(mockSessionRepo, mockUserRepo, mockTokenProvider, mockHasher)

	accessToken, refreshToken, err := uc.Login(ctx, "89991234567", "password123")

	assert.NoError(t, err)
	assert.Equal(t, "access-jwt", accessToken)
	assert.Equal(t, "refresh-jwt", refreshToken)
}

func TestValidateToken_Success(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	mToken.On("ParseToken", "valid-token").Return(int64(42), "user", nil)
	expectedUser := &domain.User{ID: 42, Name: "Илья"}
	mUser.On("GetUserByID", ctx, int64(42)).Return(expectedUser, nil)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	user, err := uc.ValidateToken(ctx, "valid-token")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestValidateToken_ParseError(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	mToken.On("ParseToken", "bad-token").Return(int64(0), "", assert.AnError)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	user, err := uc.ValidateToken(ctx, "bad-token")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Nil(t, user)
}

func TestLogin_InvalidCredentials_Phone(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	accessToken, refreshToken, err := uc.Login(ctx, "invalid-phone", "password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_UserNotFound(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	mUser.On("GetUserByPhone", ctx, "89991234567").Return(nil, infrastructure.ErrUserNotFound)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	accessToken, refreshToken, err := uc.Login(ctx, "89991234567", "password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	dbUser := &domain.User{ID: 42, Phone: "89991234567", PasswordHash: "hash"}
	mUser.On("GetUserByPhone", ctx, "89991234567").Return(dbUser, nil)
	mHash.On("CompareHashAndPassword", "hash", "wrong-password").Return(assert.AnError)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	accessToken, refreshToken, err := uc.Login(ctx, "89991234567", "wrong-password")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogout_Success(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	mSession.On("DeleteSession", ctx, "refresh-token").Return(nil)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	err := uc.Logout(ctx, "refresh-token")

	assert.NoError(t, err)
}

func TestLogout_SessionNotFound(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	mSession.On("DeleteSession", ctx, "missing-token").Return(infrastructure.ErrSessionNotFound)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	err := uc.Logout(ctx, "missing-token")

	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestRefreshTokens_Success(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	existingSession := &domain.UserSession{
		UserID:       42,
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	mSession.On("GetSessionByToken", ctx, "old-refresh").Return(existingSession, nil)

	dbUser := &domain.User{ID: 42, Role: domain.RoleUser}
	mUser.On("GetUserByID", ctx, int64(42)).Return(dbUser, nil)

	mSession.On("DeleteSession", ctx, "old-refresh").Return(nil)
	mToken.On("GenerateAccessToken", int64(42), "user").Return("new-access", nil)
	mToken.On("GenerateRefreshToken").Return("new-refresh", nil)
	mSession.On("SaveSession", ctx, mock.Anything).Return(nil)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	accessToken, refreshToken, err := uc.RefreshTokens(ctx, "old-refresh")

	assert.NoError(t, err)
	assert.Equal(t, "new-access", accessToken)
	assert.Equal(t, "new-refresh", refreshToken)
}

func TestRefreshTokens_Expired(t *testing.T) {
	ctx := context.Background()
	mUser := mocks.NewUserRepository(t)
	mSession := mocks.NewSessionRepository(t)
	mToken := mocks.NewTokenProvider(t)
	mHash := mocks.NewPasswordHasher(t)

	expiredSession := &domain.UserSession{
		UserID:       42,
		RefreshToken: "old-refresh",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}
	mSession.On("GetSessionByToken", ctx, "old-refresh").Return(expiredSession, nil)
	mSession.On("DeleteSession", ctx, "old-refresh").Return(nil)

	uc := NewAuthUsecase(mSession, mUser, mToken, mHash)
	accessToken, refreshToken, err := uc.RefreshTokens(ctx, "old-refresh")

	assert.ErrorIs(t, err, ErrSessionExpired)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}
