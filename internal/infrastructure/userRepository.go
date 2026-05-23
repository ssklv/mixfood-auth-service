package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

var userCols = []string{
	"id", "name", "phone", "email", "password_hash", "role", "created_at", "updated_at",
}

type usersRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewUserRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *usersRepository {
	return &usersRepository{db: db, psql: psql}
}

func (r *usersRepository) CreateUser(ctx context.Context, user *domain.User) error {
	builder := r.psql.Insert("users").
		Columns("name", "phone", "password_hash", "role", "created_at", "updated_at")

	values := []interface{}{user.Name, user.Phone, user.PasswordHash, user.Role, time.Now(), time.Now()}

	// Если email передан — добавляем его в запрос
	if user.Email != "" {
		builder = builder.Columns("email")
		values = append(values, user.Email)
	}

	sql, args, err := builder.Values(values...).
		Suffix("RETURNING " + strings.Join(userCols, ", ")).ToSql()
	if err != nil {
		return err
	}

	return scanUser(r.db.QueryRow(ctx, sql, args...), user)
}

func (r *usersRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	sql, args, err := r.psql.Select(userCols...).From("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, err
	}

	user := &domain.User{}
	err = scanUser(r.db.QueryRow(ctx, sql, args...), user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *usersRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	sql, args, err := r.psql.
		Select(userCols...).
		From("users").
		Where(sq.Eq{"phone": phone}).
		ToSql()
	if err != nil {
		return nil, err
	}

	user := &domain.User{}
	err = scanUser(r.db.QueryRow(ctx, sql, args...), user)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *usersRepository) UpdateUser(ctx context.Context, params *domain.UpdateUserParams) (*domain.User, error) {
	builder := r.psql.Update("users").Set("updated_at", time.Now())

	if params.Name != nil {
		builder = builder.Set("name", *params.Name)
	}
	if params.Phone != nil {
		builder = builder.Set("phone", *params.Phone)
	}
	if params.Email != nil {
		// Если передали пустую строку, возможно, стоит разрешить ставить NULL
		builder = builder.Set("email", *params.Email)
	}

	sql, args, err := builder.Where(sq.Eq{"id": params.ID}).
		Suffix("RETURNING " + strings.Join(userCols, ", ")).ToSql()
	if err != nil {
		return nil, err
	}

	user := &domain.User{}
	return user, scanUser(r.db.QueryRow(ctx, sql, args...), user)
}

func (r *usersRepository) DeleteUser(ctx context.Context, id int64) error {
	sql, args, err := r.psql.Delete("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func scanUser(row pgx.Row, user *domain.User) error {
	var email *string
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Phone,
		&email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == nil {
		if email != nil {
			user.Email = *email
		} else {
			user.Email = ""
		}
	}
	return err
}
