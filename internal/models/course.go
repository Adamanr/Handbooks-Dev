package models

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gosimple/slug"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
)

type Course struct {
	ID          uint      `db:"id"`
	Slug        string    `db:"slug"`
	Title       string    `db:"title"`
	Subtitle    string    `db:"subtitle"`
	Description string    `db:"description"`
	CoverURL    string    `db:"cover_url"`
	Status      string    `db:"status"` // draft, published, archived
	Price       float64   `db:"price"`
	Currency    string    `db:"currency"`
	Level       string    `db:"level"` // beginner, intermediate, advanced
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedByID uint      `db:"created_by_id"`

	CreatedBy User `db:"-"` // не маппится в БД
}

const (
	tableCourses = "courses"
)

var ErrorCourseNotFound = errors.New("course not found")

var courseStruct = sqlbuilder.NewStruct(Course{}).
	For(sqlbuilder.PostgreSQL)

// GetAllCourses — Получение всех курсов
func (c *Course) GetAllCourses(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) ([]Course, error) {
	sb := courseStruct.SelectFrom(tableCourses)

	sb.From(tableCourses)

	// Применяем все переданные модификаторы (фильтры, сортировки, лимиты, offset…)
	if len(opts) != 0 {
		for _, opt := range opts {
			opt(sb)
		}
	}

	sb.OrderByDesc("created_at")

	query, args := sb.Build()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot execute get all courses query",
			slog.String("query", query),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	defer rows.Close()

	courses, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Course])
	if err != nil {
		slog.ErrorContext(ctx, "cannot scan courses",
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return courses, nil
}

// GetCourseByID — получение одного курса по ID
func (c *Course) GetCourseByID(ctx context.Context, db Querier) error {
	if c.ID == 0 {
		return errors.New("course ID is required")
	}
	sb := courseStruct.SelectFrom(tableCourses)
	sb.Select(courseStruct.Columns()...)
	sb.From(tableLessons)
	sb.Where(sb.Equal("id", c.ID))
	sb.Limit(1)

	query, args := sb.Build()

	row := db.QueryRow(ctx, query, args...)

	if err := row.Scan(courseStruct.Addr(c)...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		slog.ErrorContext(ctx, "cannot get course by id",
			slog.Any("id", c.ID),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// CreateCourse — Создание курса
func (c *Course) CreateCourse(ctx context.Context, db Querier) error {
	now := time.Now()

	c.Slug = slug.Make(c.Title)
	c.CreatedAt = now
	c.UpdatedAt = now
	c.CreatedByID = 3

	sb := courseStruct.WithoutTag("db", "-").InsertInto(tableCourses, c)
	sb.SetFlavor(sqlbuilder.PostgreSQL)
	sb.Returning("id")

	query, args := sb.Build()

	var id uint
	err := db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		slog.ErrorContext(ctx, "cannot create course",
			slog.String("title", c.Title),
			slog.String("query", query),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)
		return err
	}

	c.ID = id
	return nil
}

// UpdateCourse — Изменение курса
func (c *Course) UpdateCourse(ctx context.Context, db Querier) error {
	if c.ID == 0 {
		return ErrorCourseNotFound
	}

	c.Slug = slug.Make(c.Title)
	c.UpdatedAt = time.Now()

	sb := courseStruct.WithoutTag("db", "-").Update(tableCourses, c)
	sb.Where(sb.Equal("id", c.ID))

	query, args := sb.Build()

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot update course",
			slog.Any("id", c.ID),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// DeleteCourse — Удаление курса
func (c *Course) DeleteCourse(ctx context.Context, db Querier) error {
	if c.ID == 0 {
		return ErrorCourseNotFound
	}

	sb := sqlbuilder.DeleteFrom(tableCourses).
		Where(sqlbuilder.NewCond().Equal("id", c.ID))

	query, args := sb.Build()

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot delete course",
			slog.Any("id", c.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}
