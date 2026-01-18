package models

import (
    "time"
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
