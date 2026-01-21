package models

type CourseInstructor struct {
	ID          uint
	CourseID    uint
	UserID      uint
	IsMain      bool
	Position    int
	BioOnCourse string

	Course     Course
	Instructor User
}
