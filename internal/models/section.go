package models

import (
	"time"

	"github.com/google/uuid"
)

type Section struct {
	ID            uuid.UUID `db:"id"`
	CourseID      uuid.UUID `db:"course_id"`
	CreatedID     uuid.UUID `db:"created_id" fieldtag:"immutable"`
	Title         string    `db:"title"`
	Slug          string    `db:"slug"`
	Order         int       `db:"order"`
	IsFreePreview bool      `db:"is_free_preview"`
	EstimatedTime int       `db:"estimated_time"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
