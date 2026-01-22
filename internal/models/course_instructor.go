package models

import (
	"context"
	"errors"
	"log/slog"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
)

type CourseInstructor struct {
	ID          uint   `db:"id"`
	CourseID    uint   `db:"course_id"`
	UserID      uint   `db:"course_id"`
	IsMain      bool   `db:"is_main"`
	Position    int    `db:"position"`
	BioOnCourse string `db:"bio_on_course"`

	// Не маппятся в БД, для облегчения бизнес логики
	Course     Course `db:"-"`
	Instructor User   `db:"-"`
}

const courseInstructorTable = "course_instructors"

var courseInstructorStruct = sqlbuilder.NewStruct(CourseInstructor{}).For(sqlbuilder.PostgreSQL)
var ErrorCourseInstructorNotFound = errors.New("course instructor not found")

func (c *CourseInstructor) GetCourseInstructoById(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) error {
	if c.ID == 0 {
		return errors.New("Id required")
	}

	sb := courseInstructorStruct.SelectFrom(courseInstructorTable)
	sb.Select(courseInstructorStruct.Columns()...)
	sb.From(courseInstructorTable)
	sb.Where(sb.Equal("id", c.ID))
	sb.Limit(1)

	query, args := sb.Build()

	row := db.QueryRow(ctx, query, args...)

	if err := row.Scan(courseInstructorStruct.Addr(c)...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		slog.ErrorContext(ctx, "cannot get course instructor by id",
			slog.Any("id", c.ID),
			slog.String("query", query),
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (c *CourseInstructor) GetAllCourseInstructors(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) ([]CourseInstructor, error) {
	sb := courseInstructorStruct.SelectFrom(courseInstructorTable)
	sb.From(courseInstructorTable)
	query, args := sb.Build()
	sb.OrderByDesc("is_main")

	rows, err := db.Query(ctx, query, args...)

	if err != nil {
		slog.ErrorContext(ctx, "Cant get course instructors", slog.String("query", query), slog.Any("args", args), slog.String("error", err.Error()))
		return nil, err
	}

	defer rows.Close()

	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[CourseInstructor])
	if err != nil {
		slog.ErrorContext(ctx, "Cant scan course instructors", slog.String("error", err.Error()), slog.String("query", query))
		return nil, err
	}

	return data, nil
}

func (c *CourseInstructor) CreateCoruseInstructor(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) error {
	sb := courseInstructorStruct.WithTag("db", "-").InsertInto(courseInstructorTable, c)
	sb.SetFlavor(sqlbuilder.PostgreSQL)
	sb.Returning("id")

	query, args := sb.Build()
	var id uint
	err := db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		slog.ErrorContext(ctx, "Cant insert course instructors", slog.String("query", query), slog.Any("args", args))
		return err
	}

	c.ID = id
	return nil
}

func (c *CourseInstructor) UpdateCourseInstructor(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) error {

	sb := courseInstructorStruct.WithoutTag("db", "-").Update(courseInstructorTable, c)
	sb.Where(sb.Equal("id", c.ID))

	query, args := sb.Build()

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "Cant update course instructors", slog.Any("id", c.ID), slog.String("query", query), slog.Any("args", args), slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (c *CourseInstructor) DeleteCourseInstructor(ctx context.Context, db Querier, opts ...func(*sqlbuilder.SelectBuilder)) error {

	sb := courseInstructorStruct.WithoutTag("db", "-").DeleteFrom(courseInstructorTable)
	sb.Where(sb.Equal("id", c.ID))

	query, args := sb.Build()

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "Cant delete course instructors", slog.Any("id", c.ID), slog.String("query", query), slog.Any("args", args), slog.String("error", err.Error()))
		return err
	}

	return nil
}
