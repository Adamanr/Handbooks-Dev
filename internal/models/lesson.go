package models

import (
	"time"

	"github.com/google/uuid"
)

type Lesson struct {
	ID          uuid.UUID `db:"id" fieldtag:"immutable"`
	SectionID   uuid.UUID `db:"section_id"`
	CourseID    uuid.UUID `db:"course_id"`
	CreatedID   uuid.UUID `db:"created_id"`
	Title       string    `db:"title"`
	Slug        string    `db:"slug"`
	Type        string    `db:"type"`
	Content     string    `db:"content"`
	Order       int       `db:"order"`
	DurationSec int       `db:"duration_sec"`
	IsPublished bool      `db:"is_published"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
