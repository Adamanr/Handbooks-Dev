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

// CreateLesson implements [api.ServerInterface].
func (s *Server) CreateLesson(w http.ResponseWriter, r *http.Request, courseID string, sectionID string) {
	var (
		ctx    = r.Context()
		lesson models.Lesson
	)

	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	time := time.Now()
	lesson.ID = uuid.New()
	lesson.Slug = slug.Make(lesson.Title)
	lesson.CreatedID = ctx.Value("user").(*Claims).ID
	lesson.CreatedAt = time
	lesson.UpdatedAt = time

	if err := storage.Create(ctx, "lessons", lesson, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error creating lesson", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusCreated, lesson, "lesson")
}

// GetLessons implements [api.ServerInterface].
func (s *Server) GetLessons(w http.ResponseWriter, r *http.Request, courseID string, sectionID string, params api.GetLessonsParams) {
	var ctx = r.Context()

	lessons, err := storage.GetAll[models.Lesson](ctx, "lessons", s.DB)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting lessons", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, lessons, "lessons")
}

// DeleteLesson implements [api.ServerInterface].
func (s *Server) DeleteLesson(w http.ResponseWriter, r *http.Request, courseID string, sectionID string, lessonID string) {
	var ctx = r.Context()

	if err := storage.Delete[models.Lesson](ctx, "lessons", s.DB, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", lessonID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error deleting lesson by ID", slog.String("error", err.Error()), slog.Any("ID", lessonID))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, true, "lesson")
}

// GetLessonByID implements [api.ServerInterface].
func (s *Server) GetLessonByID(w http.ResponseWriter, r *http.Request, courseID string, sectionID string, lessonID string) {
	ctx := r.Context()

	course, err := storage.GetOne[models.Lesson](ctx, s.DB, "lessons", func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", lessonID))
	})

	if err != nil {
		slog.ErrorContext(ctx, "Error getting course by id",
			slog.String("id", lessonID),
			slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}

	s.JSON(w, r, http.StatusOK, course, "course")
}

// UpdateLesson implements [api.ServerInterface].
func (s *Server) UpdateLesson(w http.ResponseWriter, r *http.Request, courseID string, sectionID string, lessonID string) {
	var (
		ctx    = r.Context()
		lesson models.Lesson
	)

	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	if err := storage.Update[models.Lesson](ctx, "lessons", lesson, s.DB, func(sb *sqlbuilder.UpdateBuilder) {
		sb.Where(sb.Equal("id", lessonID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error updating lesson", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusCreated, lesson, "lesson_id")
}

// ProgressUpdateLessonProgress implements [api.ServerInterface].
func (s *Server) ProgressUpdateLessonProgress(w http.ResponseWriter, r *http.Request, courseId string, lessonId string) {
	panic("unimplemented")
}
