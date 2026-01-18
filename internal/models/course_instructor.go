package models

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type CourseInstructor struct {
	ID			uint
	CourseID    uint
	UserID      uint
	IsMain      bool
	Position    int
	BioOnCourse string

	Course     Course
	Instructor User
}

func (c *CourseInstructor) CreateCourseInstructor(ctx context.Context, db *pgx.Conn, logger *slog.Logger) {
	if _,err:= db.QueryRow()
}
