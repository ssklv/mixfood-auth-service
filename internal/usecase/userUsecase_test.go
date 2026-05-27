package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
	"github.com/ssklv/mixfood-auth-service/internal/usecase/mocks"
)

func TestGetProfile_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	expectedUser := &domain.User{ID: 42, Name: "Илья", Phone: "89991234567"}
	mockUserRepo.On("GetUserByID", ctx, int64(42)).Return(expectedUser, nil)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	user, err := uc.GetProfile(ctx, 42)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "Илья", user.Name)
}

func TestGetProfile_UserNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	mockUserRepo.On("GetUserByID", ctx, int64(99)).Return(nil, infrastructure.ErrUserNotFound)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	user, err := uc.GetProfile(ctx, 99)

	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Nil(t, user)
}

func TestCreateAddress_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	newAddr := &domain.Address{
		UserID:      42,
		StreetHouse: "ул. Пушкина, д. 10",
		Apartment:   "12",
		Entrance:    "1",
		Floor:       "3",
		DoorCode:    "12к345",
	}
	mockAddressRepo.On("CreateAddress", ctx, newAddr).Return(nil)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.CreateAddress(ctx, newAddr)

	assert.NoError(t, err)
}

func TestCreateAddress_ValidationError_StreetHouse(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	invalidAddr := &domain.Address{
		UserID:      42,
		StreetHouse: "",
	}

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.CreateAddress(ctx, invalidAddr)

	assert.ErrorIs(t, err, ErrInvalidAddress)
}

func TestCreateAddress_ValidationError_ApartmentTooLong(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	invalidAddr := &domain.Address{
		UserID:      42,
		StreetHouse: "ул. Пушкина, д. 10",
		Apartment:   strings.Repeat("A", 25),
	}

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.CreateAddress(ctx, invalidAddr)

	assert.ErrorIs(t, err, ErrInvalidAddress)
}

func TestDeleteAddress_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	userID := int64(42)
	addressID := int64(7)

	dbAddress := &domain.Address{ID: addressID, UserID: userID, StreetHouse: "Ленина 5"}
	mockAddressRepo.On("GetAddressByID", ctx, addressID).Return(dbAddress, nil)
	mockAddressRepo.On("DeleteAddress", ctx, addressID).Return(nil)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.DeleteAddress(ctx, userID, addressID)

	assert.NoError(t, err)
}

func TestDeleteAddress_EnemyAddress_ReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	myUserID := int64(42)
	enemyUserID := int64(666)
	addressID := int64(7)

	enemyAddress := &domain.Address{ID: addressID, UserID: enemyUserID, StreetHouse: "Чужая улица 1"}
	mockAddressRepo.On("GetAddressByID", ctx, addressID).Return(enemyAddress, nil)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.DeleteAddress(ctx, myUserID, addressID)

	assert.ErrorIs(t, err, ErrAddressNotFound)
}

func TestUpdateProfile_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	name := "Новое Имя"
	phone := "89991112233"
	email := "test@test.ru"
	params := &domain.UpdateUserParams{
		Name:  &name,
		Phone: &phone,
		Email: &email,
	}

	expectedUser := &domain.User{Name: name, Phone: phone, Email: email}
	mockUserRepo.On("UpdateUser", ctx, params).Return(expectedUser, nil)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	user, err := uc.UpdateProfile(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestUpdateProfile_InvalidName(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	name := ""
	params := &domain.UpdateUserParams{Name: &name}

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	user, err := uc.UpdateProfile(ctx, params)

	assert.ErrorIs(t, err, ErrInvalidName)
	assert.Nil(t, user)
}

func TestGetAddresses_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	dbList := []domain.Address{
		{ID: 1, UserID: 42, StreetHouse: "ул. Ленина"},
	}
	mockAddressRepo.On("GetAddressesByUserID", ctx, int64(42)).Return(dbList, nil)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	list, err := uc.GetAddresses(ctx, 42)

	assert.NoError(t, err)
	assert.Equal(t, dbList, list)
}

func TestUpdateAddress_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	street := "Новая улица"
	apt := "45"
	params := &domain.UpdateAddressParams{
		ID:          1,
		StreetHouse: &street,
		Apartment:   &apt,
	}

	mockAddressRepo.On("UpdateAddress", ctx, params).Return(nil)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.UpdateAddress(ctx, int64(42), params)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), params.UserID)
}

func TestUpdateAddress_NotFound(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	street := "Новая улица"
	params := &domain.UpdateAddressParams{
		ID:          1,
		StreetHouse: &street,
	}

	mockAddressRepo.On("UpdateAddress", ctx, params).Return(infrastructure.ErrAddressNotFound)

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.UpdateAddress(ctx, int64(42), params)

	assert.ErrorIs(t, err, ErrAddressNotFound)
}

func TestUpdateAddress_ValidationError_DoorCodeTooLong(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewUserRepository(t)
	mockAddressRepo := mocks.NewAddressRepository(t)

	longCode := strings.Repeat("9", 30)
	params := &domain.UpdateAddressParams{
		ID:       1,
		DoorCode: &longCode,
	}

	uc := NewUserUsecase(mockUserRepo, mockAddressRepo)
	err := uc.UpdateAddress(ctx, int64(42), params)

	assert.ErrorIs(t, err, ErrInvalidAddress)
}
