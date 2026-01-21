package models

import (
	"context"
	"log/slog"
	"time"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
)

type Course struct {
    ID          uint
    Slug        string
    Title       string
    Subtitle    string
    Description string
    CoverURL    string
    Status      string     // draft, published, archived
    Price       float64
    Currency    string
    Level       string     // beginner, intermediate, advanced
    CreatedAt   time.Time
    UpdatedAt   time.Time

    CreatedByID uint
    CreatedBy   User
}

func (c *Course) GetAllCourses(ctx context.Context, db *pgx.Conn, logger *slog.Logger) ([]Course, error) {
	rows, err := db.Query(ctx, "SELECT * FROM courses")
	if err!=nil{
		logger.Error("cant get courses", slog.String("error", err.Error()))
		return nil, err
	}

	data,err := pgx.CollectRows(rows, pgx.RowToStructByName[Course])
	if err!=nil{
		logger.Error("cant get courses", slog.String("error", err.Error()))
		return nil, err
	}

	return data, nil
}

func (c *Course) CreateCourse(ctx context.Context, db *pgx.Conn, logger *slog.Logger) (uint, error){
	if err:=db.QueryRow(ctx, "INSERT INTO courses(slug, title, subtitle, description, cover_url, status, price, currency, level, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id", slug.Make(c.Title), c.Title,c.Subtitle,c.Description,c.CoverURL,c.Status,c.Price,c.Currency,c.Level,c.CreatedAt,c.UpdatedAt).Scan(&c.ID);err!=nil{
		logger.Error("cant create course", slog.String("error", err.Error()))
		return 0, err
	}

	return c.ID, nil
}

func (c *Course) UpdateCourse(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error {
	if _,err:=db.Exec(ctx, "UPDATE courses SET slug = $1, title = $2, subtitle = $3, description = $4, cover_url = $5, status = $6, price = $7, currency = $8, level = $9, created_at = $10, updated_at = $11 WHERE id = $12", slug.Make(c.Title), c.Title,c.Subtitle,c.Description,c.CoverURL,c.Status,c.Price,c.Currency,c.Level,c.CreatedAt,c.UpdatedAt, c.ID);err!=nil{
		logger.Error("cant update course", slog.Any("id", c.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (c *Course) DeleteCourse(ctx context.Context, db *pgx.Conn, logger *slog.Logger) error{
	if _,err := db.Exec(ctx, "DELETE FROM courses WHERE ID = $1", c.ID);err!=nil{
		logger.Error("cant delete course ", slog.Any("id", c.ID), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (c * Course)
