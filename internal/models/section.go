package models

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type Section struct {
	ID            uint
	CourseID      uint
	Title         string
	Order         int
	IsFreePreview bool
	EstimatedTime int // в минутах

	Course Course
}

func (s *Section) CreateSection(ctx context.Context, db *pgx.Conn, logger *slog.Logger) (uint, error) {
	if err := db.QueryRow(ctx, "INSERT INTO section (course_id, title, order, is_free_preview, estimated_time) values ($1, $2, $3, $4, $5) returning id", s.CourseID, s.Title, s.Order, s.IsFreePreview, s.EstimatedTime).Scan(&s.ID); err != nil {
		logger.Error("failed to create sections", slog.String("error", err.Error()))
		return 0, err
	}

	return s.ID, nil
}

func (s *Section) GetSections(ctx context.Context, db *pgx.Conn, logger *slog.Logger) ([]Section, error) {
	rows, err := db.Query(ctx, "SELECT * FROM sections")
	if err != nil {
		logger.Error("failed to get sections", slog.String("error", err.Error()))
		return nil, err
	}

	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[Section])

	if err != nil {
		logger.Error("Failed to convert sections", slog.String("error", err.Error()))
		return nil, err
	}

	return data, nil
}

func (s *Section) GetSectionById(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	row := db.QueryRow(ctx, "SELECT * FROM sections WHERE ID = $1", s.ID)

	if err := row.Scan(&s.CourseID, &s.Title, &s.Order, &s.IsFreePreview, &s.EstimatedTime); err != nil {
		logger.Error("Failed to get section", slog.Any("id", s.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *Section) UpdateSection(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	if _, err := db.Exec(ctx, "UPDATE sections SET course_id = $1, title = $2, order = $3, is_free_preview = $4, estimated_time = $5 WHERE id = $6", s.CourseID, s.Title, s.Order, s.IsFreePreview, s.EstimatedTime, s.ID); err != nil {
		logger.Error("failed to update section", slog.Any("id", s.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *Section) DeleteSection(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	if _, err := db.Exec(ctx, "DELETE FROM sections WHERE id = $1", s.ID); err != nil {
		logger.Error("failed to delete section", slog.Any("id", s.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}
