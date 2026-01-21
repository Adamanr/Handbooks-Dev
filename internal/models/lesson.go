package models

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type Lesson struct {
	ID          uint
	SectionID   uint
	Title       string
	Type        string // video, text, quiz, assignment, pdf, coding...
	Content     string
	Order       int
	DurationSec int
	IsPublished bool
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Section Section
}

func (l *Lesson) GetLessons(ctx context.Context, db *pgx.Conn, logger *slog.Logger) ([]Lesson, error) {
	rows, err := db.Query(ctx, "SELECT * FROM lessons")
	if err != nil {
		logger.Error("failed to get lessons", slog.String("error", err.Error()))
		return nil, err
	}

	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[Lesson])
	if err != nil {
		logger.Error("failed to get lessons", slog.String("error", err.Error()))
		return nil, err
	}

	return data, nil
}

func (l *Lesson) CreateLesson(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	if err := db.QueryRow(ctx, "INSERT INTO lessons (section_id, title, type, content, order, duration_sec, is_published, created_at, updated_at) values($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id", l.SectionID, l.Title, l.Type, l.Content, l.Order, l.DurationSec, l.IsPublished, l.CreatedAt, l.UpdatedAt).Scan(&l.ID); err != nil {
		logger.Error("failed to create lesson", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (l *Lesson) GetLessonById(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	row := db.QueryRow(ctx, "SELECT * FROM lessons WHERE id = $1", l.ID)

	if err := row.Scan(&l.SectionID, &l.Title, &l.Type, &l.Content, &l.Order, &l.DurationSec, &l.IsPublished, &l.CreatedAt, &l.UpdatedAt); err != nil {
		logger.Error("failed to update lesson", slog.Any("id", l.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}
