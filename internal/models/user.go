package models

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           uint
	Email        string
	PasswordHash string
	FullName     string
	AvatarURL    string
	Role         string
	CreatedAt    time.Time
	LastLoginAt  *time.Time
}

func (u *User) CreateUser(ctx context.Context, db *pgx.Conn, logger *slog.Logger) (uint, error) {
	if err := db.QueryRow(ctx, "INSERT INTO users (email, password_hash, full_name, avatar_url, role, created_at, last_login_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		u.Email, u.PasswordHash, u.FullName, u.AvatarURL, u.Role, time.Now(), time.Now()).Scan(&u.ID); err != nil {
		logger.Error("failed to create user", slog.String("email", u.Email), slog.String("error", err.Error()))
		return 0, err
	}

	return u.ID, nil
}

func (u *User) UpdateUser(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	if _, err := db.Exec(ctx, "UPDATE users SET email = $1, password_hash = $2, full_name = $3, avatar_url = $4, role = $5, last_login_at = $6 WHERE id = $7", u.Email, u.PasswordHash, u.FullName, u.AvatarURL, u.Role, time.Now(), u.ID); err != nil {
		logger.Error("failed to update user", slog.String("email", u.Email), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (u *User) DeleteUser(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	if _, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID); err != nil {
		logger.Error("failed to delete user", slog.Any("id", u.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (u *User) GetUsers(ctx context.Context, db *pgx.Conn, logger *slog.Logger) ([]User, error) {
	rows, err := db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		logger.Error("failed to get users", slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		logger.Error("Error collecting rows", slog.String("error", err.Error()))
		return nil, err
	}

	return users, nil
}

func (u *User) GetUserByID(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	row := db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", u.ID)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.LastLoginAt); err != nil {
		logger.Error("failed to get user", slog.Any("id", u.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}
