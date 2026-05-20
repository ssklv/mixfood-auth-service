package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

var userCols = []string{
	"id",
	"name",
	"phone",
	"email",
	"address",
	"password_hash",
	"role",
	"created_at",
	"updated_at",
}

type usersRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewUserRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *usersRepository {
	return &usersRepository{
		db:   db,
		psql: psql,
	}
}

func (r *usersRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	sql, args, err := r.psql.
		Select(userCols...).
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	userRes := &domain.User{}
	row := r.db.QueryRow(ctx, sql, args...)

	if err := scanUser(row, userRes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return userRes, nil
}

func (r *usersRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	sql, args, err := r.psql.
		Select(userCols...).
		From("users").
		Where(sq.Eq{"phone": phone}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("get user by phone: %w", err)
	}

	userRes := &domain.User{}
	row := r.db.QueryRow(ctx, sql, args...)

	if err := scanUser(row, userRes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return userRes, nil
}

func (r *usersRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	sql, args, err := r.psql.
		Select(userCols...).
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	userRes := &domain.User{}
	row := r.db.QueryRow(ctx, sql, args...)

	if err := scanUser(row, userRes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return userRes, nil
}

func (r *usersRepository) CreateUser(ctx context.Context, user *domain.User) error {
	sql, args, err := r.psql.
		Insert("users").
		Columns("name", "phone", "password_hash", "role", "created_at", "updated_at").
		Values(user.Name, user.Phone, user.PasswordHash, user.Role, time.Now(), time.Now()).
		Suffix("RETURNING " + strings.Join(userCols, ", ")).
		ToSql()

	if err != nil {
		return fmt.Errorf("build create user query: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	return scanUser(row, user)
}

func (r *usersRepository) UpdateUser(ctx context.Context, input *domain.UpdateUserParams) (*domain.User, error) {
	setCount := 0
	builder := r.psql.Update("users")

	if input.Name != nil {
		builder = builder.Set("name", *input.Name)
		setCount++
	}
	if input.Phone != nil {
		builder = builder.Set("phone", *input.Phone)
		setCount++
	}
	if input.Email != nil {
		builder = builder.Set("email", *input.Email)
		setCount++
	}
	if input.Address != nil {
		builder = builder.Set("address", *input.Address)
		setCount++
	}

	if setCount == 0 {
		return nil, ErrNoChanges
	}

	// Добавляем обновление времени
	builder = builder.Set("updated_at", time.Now())

	sql, args, err := builder.
		Where(sq.Eq{"id": input.ID}).
		Suffix("RETURNING " + strings.Join(userCols, ", ")).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build update query: %w", err)
	}

	user := &domain.User{}
	row := r.db.QueryRow(ctx, sql, args...)

	if err := scanUser(row, user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "phone") {
				return nil, ErrDuplicatePhone
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return nil, ErrDuplicateEmail
			}
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (r *usersRepository) DeleteUser(ctx context.Context, id int64) error {
	sql, args, err := r.psql.
		Delete("users").
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func scanUser(row pgx.Row, user *domain.User) error {
	return row.Scan(
		&user.ID,
		&user.Name,
		&user.Phone,
		&user.Email,
		&user.Address,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}
