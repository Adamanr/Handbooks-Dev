package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// NewDatabase - Functions for create new connections for database
func NewDatabase(ctx context.Context, url string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to connect to database", slog.String("Error", err.Error()))
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		slog.ErrorContext(ctx, "Error database ping", slog.String("Error", err.Error()))
		return nil, err
	}

	return conn, nil
}
