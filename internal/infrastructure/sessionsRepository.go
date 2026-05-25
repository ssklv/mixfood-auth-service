package infrastructure

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssklv/mixfood-auth-service/internal/domain"
)

type SessionsRepository struct {
	db   *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewSessionRepository(db *pgxpool.Pool, psql sq.StatementBuilderType) *SessionsRepository {
	return &SessionsRepository{db: db, psql: psql}
}

func (r *SessionsRepository) SaveSession(ctx context.Context, session *domain.UserSession) error {
	sql, args, err := r.psql.
		Insert("sessions").
		Columns("user_id", "refresh_token", "expires_at").
		Values(session.UserID, session.RefreshToken, session.ExpiresAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return ErrDatabaseInternal
	}
	return nil
}

func (r *SessionsRepository) GetSessionByToken(ctx context.Context, token string) (*domain.UserSession, error) {
	sql, args, err := r.psql.
		Select("user_id", "refresh_token", "expires_at").
		From("sessions").
		Where(sq.Eq{"refresh_token": token}).
		ToSql()
	if err != nil {
		return nil, err
	}
	s := &domain.UserSession{}
	err = r.db.QueryRow(ctx, sql, args...).Scan(&s.UserID, &s.RefreshToken, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, ErrDatabaseInternal
	}
	return s, nil
}

func (r *SessionsRepository) DeleteSession(ctx context.Context, token string) error {
	sql, args, err := r.psql.
		Delete("sessions").
		Where(sq.Eq{"refresh_token": token}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return ErrDatabaseInternal
	}
	return nil
}
