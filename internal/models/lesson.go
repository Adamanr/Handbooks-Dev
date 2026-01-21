package models

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
)

type Lesson struct {
	ID          uint      `db:"id"`
	SectionID   uint      `db:"section_id"`
	Title       string    `db:"title"`
	Type        string    `db:"type"` // video, text, quiz, assignment, pdf, coding...
	Content     string    `db:"content"`
	Order       int       `db:"order"`
	DurationSec int       `db:"duration_sec"`
	IsPublished bool      `db:"is_published"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	// Не маппится в БД, только для удобства в бизнес-логике
	Section Section `db:"-"`
}

const tableLessons = "lessons"

var ErrNotFound = errors.New("lesson not found")

var lessonStruct = sqlbuilder.NewStruct(Lesson{}).
	For(sqlbuilder.PostgreSQL)

// GetLessons — получение списка уроков с поддержкой динамических фильтров
func (l *Lesson) GetLessons(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) ([]Lesson, error) {
	sb := lessonStruct.SelectFrom(tableLessons)
	sb.From(tableLessons)

	// Применяем все опции (фильтры, сортировки, лимиты, offset и т.д.)
	if len(opts) != 0 {
		for _, opt := range opts {
			opt(sb)
		}
	}

	sb.OrderByDesc("created_at")

	query, args := sb.Build()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot fetch lessons",
			slog.String("query", query),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	defer rows.Close()

	lessons, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Lesson])
	if err != nil {
		slog.ErrorContext(ctx, "cannot scan lessons",
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return lessons, nil
}

// GetLessonByID — получение одного урока по ID
func (l *Lesson) GetLessonByID(ctx context.Context, db Querier) error {
	if l.ID == 0 {
		return errors.New("lesson ID is required")
	}

	sb := lessonStruct.SelectFrom(tableLessons)
	sb.Select(lessonStruct.Columns()...)
	sb.From(tableLessons)
	sb.Where(sb.Equal("id", l.ID))
	sb.Limit(1)

	query, args := sb.Build()

	row := db.QueryRow(ctx, query, args...)

	if err := row.Scan(lessonStruct.Addr(l)...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		slog.ErrorContext(ctx, "cannot get lesson by id",
			slog.Any("id", l.ID),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// CreateLesson — создание нового урока
func (l *Lesson) CreateLesson(ctx context.Context, db Querier) error {
	now := time.Now()
	l.CreatedAt = now
	l.UpdatedAt = now

	// Если порядок не задан — можно задать дефолт (опционально)
	if l.Order == 0 {
		l.Order = 1 // или логика вычисления следующего порядка в разделе
	}

	sb := lessonStruct.WithoutTag("db", "-").InsertInto(tableLessons, l)
	sb.Returning("id")

	query, args := sb.Build()

	err := db.QueryRow(ctx, query, args...).Scan(&l.ID)
	if err != nil {
		slog.ErrorContext(ctx, "cannot create lesson",
			slog.Any("section_id", l.SectionID),
			slog.String("title", l.Title),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

// UpdateLesson — обновление урока
func (l *Lesson) UpdateLesson(ctx context.Context, db Querier) error {
	if l.ID == 0 {
		return errors.New("lesson ID is required")
	}

	l.UpdatedAt = time.Now()

	sb := lessonStruct.WithoutTag("db", "-").Update(tableLessons, l)
	sb.Where(sb.Equal("id", l.ID))

	query, args := sb.Build()

	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot update lesson",
			slog.Any("id", l.ID),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteLesson — удаление урока
func (l *Lesson) DeleteLesson(ctx context.Context, db Querier) error {
	if l.ID == 0 {
		return errors.New("lesson ID is required")
	}

	sb := sqlbuilder.DeleteFrom(tableLessons).
		Where(sqlbuilder.NewCond().Equal("id", l.ID))

	query, args := sb.Build()

	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot delete lesson",
			slog.Any("id", l.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
