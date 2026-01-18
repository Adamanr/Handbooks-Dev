package models

type Section struct {
	ID            uint
	CourseID      uint
	Title         string
	Order         int
	IsFreePreview bool
	EstimatedTime int // в минутах

	Course Course
}
