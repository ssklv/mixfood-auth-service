package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect просто открывает соединение с базой данных
func Connect(connectionString string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), connectionString)
}
