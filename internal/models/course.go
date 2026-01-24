package models

import (
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID          string    `db:"id" fieldtag:"immutable"`
	Slug        string    `db:"slug"`
	Title       string    `db:"title"`
	Subtitle    string    `db:"subtitle"`
	Description string    `db:"description"`
	CoverURL    string    `db:"cover_url"`
	Status      string    `db:"status"`
	Price       float64   `db:"price"`
	Currency    string    `db:"currency"`
	Level       string    `db:"level"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedID   uuid.UUID `db:"created_id" fieldtag:"immutable"`
}
