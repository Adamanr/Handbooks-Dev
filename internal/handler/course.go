package handlers

import (
	"encoding/json"
	"handbooks/internal/api"
	"handbooks/internal/models"
	storage "handbooks/pkg/storage"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/huandu/go-sqlbuilder"
)

// CoursesCreate implements [api.ServerInterface].
func (s *Server) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		course models.Course
	)

	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	time := time.Now()
	course.ID = uuid.New().String()
	course.Slug = slug.Make(course.Title)
	course.CreatedID = ctx.Value("user").(*Claims).ID
	course.CreatedAt = time
	course.UpdatedAt = time

	if err := storage.Create(ctx, "courses", course, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error creating course", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusCreated, course.ID, "course")
}

// CoursesList implements [api.ServerInterface].
func (s *Server) GetCourses(w http.ResponseWriter, r *http.Request, params api.GetCoursesParams) {
	var ctx = r.Context()

	courses, err := storage.GetAll[models.Course](ctx, "courses", s.DB)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting courses", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, courses, "courses")
}

// CoursesGetById implements [api.ServerInterface].
func (s *Server) GetCourseByID(w http.ResponseWriter, r *http.Request, courseID string) {
	ctx := r.Context()

	course, err := storage.GetOne[models.Course](ctx, s.DB, "courses", func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", courseID))
	})

	if err != nil {
		slog.ErrorContext(ctx, "Error getting course by ID",
			slog.String("id", courseID),
			slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}

	s.JSON(w, r, http.StatusOK, course, "course")
}

// UpdateCourse implements [api.ServerInterface].
func (s *Server) UpdateCourse(w http.ResponseWriter, r *http.Request, courseID string) {
	var (
		ctx    = r.Context()
		course models.Course
	)

	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	if err := storage.Update[models.Course](ctx, "courses", course, s.DB, func(sb *sqlbuilder.UpdateBuilder) {
		sb.Where(sb.Equal("id", courseID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error updating course", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusCreated, course.ID, "course_id")
}

// DeleteCourse implements [api.ServerInterface].
func (s *Server) DeleteCourse(w http.ResponseWriter, r *http.Request, courseID string) {
	var ctx = r.Context()

	if err := storage.Delete[models.Course](ctx, "courses", s.DB, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", courseID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error deleting course by id", slog.String("error", err.Error()), slog.Any("ID", courseID))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, true, "course")
}
