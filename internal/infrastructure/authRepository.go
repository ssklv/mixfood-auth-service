package infrastructure

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

type authRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewAuthRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *authRepository {
	return &authRepository{
		db:   db,
		psql: psql,
	}
}

/////

func (r *authRepository) CreateUser(ctx context.Context, user *domain.User) error {
	sql, args, err := r.psql.
		Insert("users").
		Columns("name", "phone", "password_hash", "role").
		Values(user.Name, user.Phone, user.PasswordHash, user.Role).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		return fmt.Errorf("build create query: %w", err)
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicatePhone
		}
		return fmt.Errorf("execute create user: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	sql, args, err := r.psql.
		Select("id", "name", "phone", "password_hash", "role", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"phone": phone}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build select query: %w", err)
	}

	user := &domain.User{}
	err = r.db.QueryRow(ctx, sql, args...).
		Scan(
			&user.ID, &user.Name, &user.Phone, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return user, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	sql, args, err := r.psql.
		Select("id", "name", "phone", "password_hash", "role", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build select by id query: %w", err)
	}

	user := &domain.User{}
	err = r.db.QueryRow(ctx, sql, args...).
		Scan(
			&user.ID, &user.Name, &user.Phone, &user.PasswordHash,
			&user.Role, &user.CreatedAt, &user.UpdatedAt,
		)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user by id: %w", err)
	}
	return user, nil
}

func (r *authRepository) GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error) {
	sql, args, err := r.psql.Select("user_id", "refresh_token", "expires_at").
		From("sessions").
		Where(sq.Eq{"refresh_token": token}).
		ToSql()
	if err != nil {
		return nil, err
	}

	session := &domain.UserSession{}
	err = r.db.QueryRow(ctx, sql, args...).
		Scan(&session.UserID, &session.RefreshToken, &session.ExpiresAt)
	return session, err
}

func (r *authRepository) SaveSession(ctx context.Context, session *domain.UserSession) error {
	sql, args, err := r.psql.
		Insert("sessions").
		Columns("user_id", "refresh_token", "expires_at").
		Values(session.UserID, session.RefreshToken, session.ExpiresAt).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET refresh_token = EXCLUDED.refresh_token, expires_at = EXCLUDED.expires_at").
		ToSql()

	if err != nil {
		return fmt.Errorf("build save session query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("execute save session: %w", err)
	}
	return nil
}

func (r *authRepository) DeleteSession(ctx context.Context, refreshToken string) error {
	sql, args, err := r.psql.
		Delete("sessions").
		Where(sq.Eq{"refresh_token": refreshToken}).
		ToSql()

	if err != nil {
		return fmt.Errorf("build delete session query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}
