package models

type CourseInstructor struct {
	CourseID    uint
	UserID      uint
	IsMain      bool
	Position    int
	BioOnCourse string `gorm:"type:text"`

	Course     Course `gorm:"constraint:OnDelete:CASCADE"`
	Instructor User   `gorm:"constraint:OnDelete:CASCADE"`
}
