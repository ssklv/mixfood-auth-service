package infrastructure

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

var addressCols = []string{
	"id",
	"user_id",
	"street_house",
	"apartment",
	"entrance",
	"floor",
	"door_code",
}

type AddressRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewAddressRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *AddressRepository {
	return &AddressRepository{db: db, psql: psql}
}

func (r *AddressRepository) CreateAddress(ctx context.Context, addr *domain.Address) error {
	sql, args, err := r.psql.
		Insert("addresses").
		Columns("user_id", "street_house", "apartment", "entrance", "floor", "door_code").
		Values(addr.UserID, addr.StreetHouse, addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, sql, args...).Scan(&addr.ID)
}

func (r *AddressRepository) GetAddressByID(ctx context.Context, id int64) (*domain.Address, error) {
	sql, args, err := r.psql.
		Select(addressCols...).
		From("addresses").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var a domain.Address
	err = scanAddress(r.db.QueryRow(ctx, sql, args...), &a)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAddressNotFound
		}
		return nil, ErrDatabaseInternal
	}
	return &a, nil
}

func (r *AddressRepository) GetAddressesByUserID(ctx context.Context, userID int64) ([]domain.Address, error) {
	sql, args, err := r.psql.
		Select(addressCols...).
		From("addresses").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addrs := make([]domain.Address, 0)
	for rows.Next() {
		var a domain.Address
		if err := scanAddress(rows, &a); err != nil {
			return nil, err
		}
		addrs = append(addrs, a)
	}
	return addrs, nil
}

func (r *AddressRepository) DeleteAddress(ctx context.Context, id int64) error {
	sql, args, err := r.psql.
		Delete("addresses").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return ErrDatabaseInternal
	}
	if res.RowsAffected() == 0 {
		return ErrAddressNotFound
	}
	return nil
}

func (r *AddressRepository) UpdateAddress(ctx context.Context, params domain.UpdateAddressParams) error {
	builder := r.psql.
		Update("addresses").
		Where(sq.Eq{"id": params.ID, "user_id": params.UserID})

	if params.StreetHouse != nil {
		builder = builder.Set("street_house", *params.StreetHouse)
	}
	if params.Apartment != nil {
		builder = builder.Set("apartment", *params.Apartment)
	}
	if params.Entrance != nil {
		builder = builder.Set("entrance", *params.Entrance)
	}
	if params.Floor != nil {
		builder = builder.Set("floor", *params.Floor)
	}
	if params.DoorCode != nil {
		builder = builder.Set("door_code", *params.DoorCode)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return ErrDatabaseInternal
	}
	if res.RowsAffected() == 0 {
		return ErrAddressNotFound
	}
	return nil
}
func scanAddress(row pgx.Row, addr *domain.Address) error {
	return row.Scan(
		&addr.ID,
		&addr.UserID,
		&addr.StreetHouse,
		&addr.Apartment,
		&addr.Entrance,
		&addr.Floor,
		&addr.DoorCode,
	)
}
