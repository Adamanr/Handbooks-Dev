package models

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID          uint       `db:"id"`
	Email       string     `db:"email"`
	FullName    string     `db:"full_name"`
	AvatarURL   string     `db:"avatar_url"`
	Role        string     `db:"role"` // student, instructor, admin
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	LastLoginAt *time.Time `db:"last_login_at"`

	// Не сохраняем в БД и не читаем из БД после загрузки (безопасность)
	PasswordHash string `db:"-"`
}

const tableUsers = "users"

var userStruct = sqlbuilder.NewStruct(User{}).
	For(sqlbuilder.PostgreSQL)

// CreateUser — регистрация нового пользователя
func (u *User) CreateUser(ctx context.Context, db Querier) (uint, error) {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	sb := userStruct.WithoutTag("db", "-").InsertInto(tableUsers, u)
	sb.Returning("id")

	query, args := sb.Build()

	err := db.QueryRow(ctx, query, args...).Scan(&u.ID)
	if err != nil {
		slog.ErrorContext(ctx, "cannot create user",
			slog.String("email", u.Email),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return 0, err // можно добавить проверку на unique violation
	}

	return u.ID, nil
}

// GetUserByEmail — основной метод для логина
func (u *User) GetUserByEmail(ctx context.Context, db Querier) error {
	sb := userStruct.SelectFrom(tableUsers)
	sb.From(tableUsers)
	sb.Where(sb.Equal("email", u.Email))
	sb.Limit(1)

	query, args := sb.Build()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot query user by email",
			slog.String("email", u.Email),
			slog.String("error", err.Error()),
		)
		return err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[User])
	if err != nil {
		slog.ErrorContext(ctx, "cannot scan user by email",
			slog.String("email", u.Email),
			slog.String("error", err.Error()),
		)
		return err
	}

	if len(users) == 0 {
		return ErrNotFound
	}

	*u = users[0]
	return nil
}

// GetUserByID — получение пользователя по ID
func (u *User) GetUserByID(ctx context.Context, db Querier) error {
	if u.ID == 0 {
		return errors.New("user ID is required")
	}

	sb := userStruct.SelectFrom(tableUsers)
	sb.From(tableUsers)
	sb.Where(sb.Equal("id", u.ID))
	sb.Limit(1)

	query, args := sb.Build()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[User])
	if err != nil {
		slog.ErrorContext(ctx, "cannot scan user by id",
			slog.Any("id", u.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	if len(users) == 0 {
		return ErrNotFound
	}

	*u = users[0]
	return nil
}

// UpdateUser — обновление профиля (без смены пароля)
func (u *User) UpdateUser(ctx context.Context, db Querier) error {
	if u.ID == 0 {
		return errors.New("user ID is required")
	}

	u.UpdatedAt = time.Now()

	sb := userStruct.WithoutTag("db", "-").Update(tableUsers, u)
	sb.Where(sb.Equal("id", u.ID))

	query, args := sb.Build()

	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot update user",
			slog.Any("id", u.ID),
			slog.String("email", u.Email),
			slog.String("error", err.Error()),
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdatePassword — отдельный метод для смены пароля (безопаснее)
func (u *User) UpdatePassword(ctx context.Context, db Querier, newHash string) error {
	if u.ID == 0 {
		return errors.New("user ID is required")
	}

	sb := sqlbuilder.Update(tableUsers).
		Set("password_hash", newHash).
		Set("updated_at", time.Now().String()).
		Where(sqlbuilder.NewCond().Equal("id", u.ID))

	query, args := sb.Build()

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot update user password",
			slog.Any("id", u.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	u.PasswordHash = "" // очищаем в памяти
	return nil
}

// DeleteUser — удаление пользователя
func (u *User) DeleteUser(ctx context.Context, db Querier) error {
	if u.ID == 0 {
		return errors.New("user ID is required")
	}

	sb := sqlbuilder.DeleteFrom(tableUsers).
		Where(sqlbuilder.NewCond().Equal("id", u.ID))

	query, args := sb.Build()

	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot delete user",
			slog.Any("id", u.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetUsers — список пользователей (для админов, с пагинацией/фильтрами)
func (u *User) GetUsers(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) ([]User, error) {
	sb := userStruct.SelectFrom(tableUsers)
	sb.From(tableUsers)

	for _, opt := range opts {
		opt(sb)
	}

	sb.OrderByAsc("created_at")

	query, args := sb.Build()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot fetch users",
			slog.String("query", query),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[User])
	if err != nil {
		slog.ErrorContext(ctx, "cannot scan users",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return users, nil
}
