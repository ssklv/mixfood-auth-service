package usecase

import (
	"context"
	"testing"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

type mockAuthRepository struct {
	errToReturn     error
	userToReturn    *domain.User
	sessionToReturn *domain.UserSession
}

func (mr *mockAuthRepository) CreateUser(ctx context.Context, user *domain.User) error {
	return mr.errToReturn
}

func (mr *mockAuthRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return mr.userToReturn, mr.errToReturn
}

func (mr *mockAuthRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return mr.userToReturn, mr.errToReturn
}

func (mr *mockAuthRepository) SaveSession(ctx context.Context, user *domain.UserSession) error {
	return mr.errToReturn
}

func (mr *mockAuthRepository) GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error) {
	return mr.sessionToReturn, mr.errToReturn
}

func (mr *mockAuthRepository) DeleteSession(ctx context.Context, refreshToken string) error {
	return mr.errToReturn
}

// /
type mockTokenProvider struct {
	errToReturn error
}

func (mtp *mockTokenProvider) GenerateAccessToken(userID int64, role string) (string, error) {
	return "access_token", mtp.errToReturn
}

func (mtp *mockTokenProvider) GenerateRefreshToken() (string, error) {
	return "refresh_token", mtp.errToReturn
}
func (mtp *mockTokenProvider) ParseToken(tokenString string) (int64, string, error) {
	return 1, "parse_token", mtp.errToReturn
}

// /
type mockHasher struct {
	errToReturn error
}

func (mh *mockHasher) HashPassword(password string) (string, error) {
	return "hasher_password", mh.errToReturn
}
func (mh *mockHasher) CompareHashAndPassword(hash, password string) error {
	return mh.errToReturn
}

// Register(ctx context.Context, phone, password, name string) (string, string, error)
// Login(ctx context.Context, phone, password string) (string, string, error)
// Logout(ctx context.Context, refreshToken string) error
// ValidateToken(ctx context.Context, tokenString string) (*domain.User, error)
// RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
// GetUserByID(ctx context.Context, id int64) (*domain.User, error)

func Test_Register(t *testing.T) {
	mockAuthRepository := &mockAuthRepository{}
	mockTokenProvider := &mockTokenProvider{}
	mockHasher := &mockHasher{}
	au := NewAuthUsecase(mockAuthRepository, mockTokenProvider, mockHasher)

	acc, ref, err := au.Register(context.Background(), "71234567890", "passwordtest", "Иван")
	if err != nil {
		t.Errorf("")
	}
	if acc == "" || ref == "" {
		t.Errorf("")
	}
}

func Test_Login_Success(t *testing.T) {
	mockRepo := &mockAuthRepository{
		userToReturn: &domain.User{
			ID:           1,
			PasswordHash: "any_hash",
		},
	}
	mockHasher := &mockHasher{
		errToReturn: nil}
	mockTokens := &mockTokenProvider{}
	au := NewAuthUsecase(mockRepo, mockTokens, mockHasher)

	acc, ref, err := au.Login(context.Background(), "79990000000", "password123")
	if err != nil {
		t.Errorf("%v", err)
	}
	if acc == "" || ref == "" {
		t.Errorf("")
	}

}
