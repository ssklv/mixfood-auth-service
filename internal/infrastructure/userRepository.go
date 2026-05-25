package infrastructure

import (
	"context"
	"errors"
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
	"password_hash",
	"role",
	"created_at",
	"updated_at",
}

type UsersRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewUserRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *UsersRepository {
	return &UsersRepository{db: db, psql: psql}
}

func (r *UsersRepository) CreateUser(ctx context.Context, user *domain.User) error {
	columns := []string{"name", "phone", "password_hash", "role", "created_at", "updated_at"}
	values := []interface{}{user.Name, user.Phone, user.PasswordHash, user.Role, time.Now(), time.Now()}

	if user.Email != "" {
		columns = append(columns, "email")
		values = append(values, user.Email)
	}

	sql, args, err := r.psql.
		Insert("users").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING " + strings.Join(userCols, ", ")).
		ToSql()
	if err != nil {
		return err
	}

	err = scanUser(r.db.QueryRow(ctx, sql, args...), user)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 — код ошибки Unique Violation в PostgreSQL
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			errStr := strings.ToLower(pgErr.Message + " " + pgErr.ConstraintName)
			if strings.Contains(errStr, "phone") {
				return ErrDuplicatePhone
			}
			if strings.Contains(errStr, "email") {
				return ErrDuplicateEmail
			}
		}
		return ErrDatabaseInternal
	}
	return nil
}

func (r *UsersRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
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
		return nil, ErrDatabaseInternal
	}
	return user, nil
}

func (r *UsersRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	sql, args, err := r.psql.Select(userCols...).From("users").Where(sq.Eq{"phone": phone}).ToSql()
	if err != nil {
		return nil, err
	}

	user := &domain.User{}
	err = scanUser(r.db.QueryRow(ctx, sql, args...), user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, ErrDatabaseInternal
	}
	return user, nil
}

func (r *UsersRepository) UpdateUser(ctx context.Context, params *domain.UpdateUserParams) (*domain.User, error) {
	builder := r.psql.Update("users").Set("updated_at", time.Now())

	if params.Name != nil {
		builder = builder.Set("name", *params.Name)
	}
	if params.Phone != nil {
		builder = builder.Set("phone", *params.Phone)
	}
	if params.Email != nil {
		builder = builder.Set("email", *params.Email)
	}

	sql, args, err := builder.Where(sq.Eq{"id": params.ID}).Suffix("RETURNING " + strings.Join(userCols, ", ")).ToSql()
	if err != nil {
		return nil, err
	}

	user := &domain.User{}
	err = scanUser(r.db.QueryRow(ctx, sql, args...), user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, ErrDatabaseInternal
	}
	return user, nil
}

func (r *UsersRepository) DeleteUser(ctx context.Context, id int64) error {
	sql, args, err := r.psql.Delete("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return ErrDatabaseInternal
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
