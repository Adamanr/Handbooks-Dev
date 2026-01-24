package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
)

const (
	pgDriverName    = "pgx"
	migrationsDir   = "./migrations"
	gooseDriverName = "postgres"
)

// NewDatabase создает новое подключение к базе данных
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

// RunMigrations запускает миграции базы данных
func RunMigrations(ctx context.Context, dbURL string) error {
	db, err := sql.Open(pgDriverName, dbURL)
	if err != nil {
		return fmt.Errorf("не удалось открыть соединение для миграций: %w", err)
	}
	defer db.Close()

	goose.SetDialect(gooseDriverName)

	statusErr := goose.Status(db, migrationsDir)
	if statusErr != nil {
		slog.WarnContext(ctx, "Не удалось получить статус миграций", "error", statusErr)
	}

	current, _ := goose.GetDBVersion(db)
	slog.InfoContext(ctx, "Текущая версия БД", "version", current)

	slog.DebugContext(ctx, "Запуск миграций Goose...")
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}

	slog.InfoContext(ctx, "Миграции успешно применены или уже актуальны")
	return nil
}
