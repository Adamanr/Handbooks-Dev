package models

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
)

type Section struct {
	ID            uint      `db:"id"`
	CourseID      uint      `db:"course_id"`
	Title         string    `db:"title"`
	Order         int       `db:"order"`
	IsFreePreview bool      `db:"is_free_preview"`
	EstimatedTime int       `db:"estimated_time"` // в минутах
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`

	// Не маппится в БД — только для бизнес-логики
	Course Course `db:"-"`
}

const tableSections = "section"

var sectionStruct = sqlbuilder.NewStruct(Section{}).
	For(sqlbuilder.PostgreSQL)

// GetSections — получение списка разделов с поддержкой динамических фильтров
func (s *Section) GetSections(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) ([]Section, error) {
	sb := sectionStruct.SelectFrom(tableSections)
	sb.From(tableSections)

	// Применяем все переданные опции (фильтры, сортировки, лимиты...)
	if len(opts) != 0 {
		for _, opt := range opts {
			opt(sb)
		}
	}

	sb.OrderByDesc("created_at")

	query, args := sb.Build()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot fetch sections",
			slog.String("query", query),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	defer rows.Close()

	sections, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Section])
	if err != nil {
		slog.ErrorContext(ctx, "cannot scan sections",
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return sections, nil
}

// GetSectionByID — получение одного раздела по ID
func (s *Section) GetSectionByID(ctx context.Context, db Querier) error {
	if s.ID == 0 {
		return errors.New("section ID is required")
	}

	sb := sectionStruct.SelectFrom(tableSections)
	sb.From(tableSections)
	sb.Where(sb.Equal("id", s.ID))
	sb.Limit(1)

	query, args := sb.Build()

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot execute query for section by id",
			slog.Any("id", s.ID),
			slog.String("error", err.Error()),
		)
		return err
	}
	defer rows.Close()

	sections, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Section])
	if err != nil {
		slog.ErrorContext(ctx, "cannot scan section by id",
			slog.Any("id", s.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	if len(sections) == 0 {
		return ErrNotFound
	}

	*s = sections[0]
	return nil
}

// CreateSection — создание нового раздела
func (s *Section) CreateSection(ctx context.Context, db Querier) (uint, error) {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	// Если порядок не задан — можно поставить дефолтный
	if s.Order == 0 {
		s.Order = 1 // или логика вычисления следующего порядка в курсе
	}

	sb := sectionStruct.WithoutTag("db", "-").InsertInto(tableSections, s)
	sb.Returning("id")

	query, args := sb.Build()

	err := db.QueryRow(ctx, query, args...).Scan(&s.ID)
	if err != nil {
		slog.ErrorContext(ctx, "cannot create section",
			slog.Any("course_id", s.CourseID),
			slog.String("title", s.Title),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return 0, err
	}

	return s.ID, nil
}

// UpdateSection — обновление раздела
func (s *Section) UpdateSection(ctx context.Context, db Querier) error {
	if s.ID == 0 {
		return errors.New("section ID is required")
	}

	s.UpdatedAt = time.Now()

	sb := sectionStruct.WithoutTag("db", "-").Update(tableSections, s)
	sb.Where(sb.Equal("id", s.ID))

	query, args := sb.Build()

	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot update section",
			slog.Any("id", s.ID),
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

// DeleteSection — удаление раздела
func (s *Section) DeleteSection(ctx context.Context, db Querier) error {
	if s.ID == 0 {
		return errors.New("section ID is required")
	}

	sb := sqlbuilder.DeleteFrom(tableSections).
		Where(sqlbuilder.NewCond().Equal("id", s.ID))

	query, args := sb.Build()

	cmd, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "cannot delete section",
			slog.Any("id", s.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
