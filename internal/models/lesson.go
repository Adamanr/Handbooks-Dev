package models

import "time"

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
