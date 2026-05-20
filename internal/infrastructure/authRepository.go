package infrastructure

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	Users    *usersRepository
	Sessions *sessionsRepository
	Address  *addressRepository
}

func NewAuthRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *AuthRepository {
	return &AuthRepository{
		Users:    NewUserRepository(db, psql),
		Sessions: NewSessionRepository(db, psql),
		Address:  NewAddressRepository(db, psql),
	}
}
