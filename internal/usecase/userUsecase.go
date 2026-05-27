package usecase

import (
	"context"
	"errors"

	"github.com/ssklv/mixfood-auth-service/internal/domain"
	"github.com/ssklv/mixfood-auth-service/internal/infrastructure"
)

type userUsecase struct {
	userRepo    UserRepository
	addressRepo AddressRepository
}

func NewUserUsecase(userRepo UserRepository, addressRepo AddressRepository) UserUsecase {
	return &userUsecase{
		userRepo:    userRepo,
		addressRepo: addressRepo,
	}
}

func (uu *userUsecase) GetProfile(ctx context.Context, id int64) (*domain.User, error) {
	user, err := uu.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, ErrInternal
	}
	return user, nil
}

func (uu *userUsecase) UpdateProfile(ctx context.Context, params *domain.UpdateUserParams) (*domain.User, error) {
	if params.Name != nil {
		if err := validateName(*params.Name); err != nil {
			return nil, err
		}
	}
	if params.Phone != nil {
		if err := validatePhone(*params.Phone); err != nil {
			return nil, err
		}
	}
	if params.Email != nil {
		if err := validateEmail(*params.Email); err != nil {
			return nil, err
		}
	}

	user, err := uu.userRepo.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		if errors.Is(err, infrastructure.ErrDuplicateEmail) {
			return nil, ErrInvalidEmail
		}
		return nil, ErrInternal
	}
	return user, nil
}

func (uu *userUsecase) CreateAddress(ctx context.Context, addr *domain.Address) error {
	if err := validateStreetHouse(addr.StreetHouse); err != nil {
		return err
	}
	if err := validateApartment(addr.Apartment); err != nil {
		return err
	}
	if err := validateEntrance(addr.Entrance); err != nil {
		return err
	}
	if err := validateFloor(addr.Floor); err != nil {
		return err
	}
	if err := validateDoorCode(addr.DoorCode); err != nil {
		return err
	}

	err := uu.addressRepo.CreateAddress(ctx, addr)
	if err != nil {
		return ErrInternal
	}
	return nil
}

func (uu *userUsecase) GetAddresses(ctx context.Context, userID int64) ([]domain.Address, error) {
	addrs, err := uu.addressRepo.GetAddressesByUserID(ctx, userID)
	if err != nil {
		return nil, ErrInternal
	}
	return addrs, nil
}

func (uu *userUsecase) UpdateAddress(ctx context.Context, userID int64, params *domain.UpdateAddressParams) error {
	if params.StreetHouse != nil {
		if err := validateStreetHouse(*params.StreetHouse); err != nil {
			return err
		}
	}
	if params.Apartment != nil {
		if err := validateApartment(*params.Apartment); err != nil {
			return err
		}
	}
	if params.Entrance != nil {
		if err := validateEntrance(*params.Entrance); err != nil {
			return err
		}
	}
	if params.Floor != nil {
		if err := validateFloor(*params.Floor); err != nil {
			return err
		}
	}
	if params.DoorCode != nil {
		if err := validateDoorCode(*params.DoorCode); err != nil {
			return err
		}
	}

	params.UserID = userID

	err := uu.addressRepo.UpdateAddress(ctx, *params)
	if err != nil {
		if errors.Is(err, infrastructure.ErrAddressNotFound) {
			return ErrAddressNotFound
		}
		return ErrInternal
	}
	return nil
}

func (uu *userUsecase) DeleteAddress(ctx context.Context, userID int64, addressID int64) error {
	addr, err := uu.addressRepo.GetAddressByID(ctx, addressID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrAddressNotFound) {
			return ErrAddressNotFound
		}
		return ErrInternal
	}

	if addr.UserID != userID {
		return ErrAddressNotFound
	}

	err = uu.addressRepo.DeleteAddress(ctx, addressID)
	if err != nil {
		return ErrInternal
	}
	return nil
}
