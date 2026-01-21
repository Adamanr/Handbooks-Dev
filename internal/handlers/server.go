package handlers

import (
	"encoding/json"
	"handbooks/internal/api"
	"handbooks/internal/config"
	"handbooks/internal/models"
	"handbooks/pkg/requestid"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	slogchi "github.com/samber/slog-chi"
)

type Server struct {
	DB     *pgx.Conn
	Config *config.Config
}

// NewServer - functions for return server object
func NewServer(db *pgx.Conn, config *config.Config) *Server {
	return &Server{
		DB:     db,
		Config: config,
	}
}

// Run - functions for run http Server with settings
func (s *Server) Run() {
	r := chi.NewMux()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(slogchi.NewWithConfig(slog.Default(), slogchi.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,  // 400–499 → Warn
		ServerErrorLevel: slog.LevelError, // 500+   → Error
		WithRequestID:    true,            // берёт request-id из контекста
		Filters: []slogchi.Filter{
			slogchi.IgnorePath("/health", "/metrics", "/favicon.ico"),
		},
	}))

	r.Use(requestid.MiddlewareRequestID)

	h := api.HandlerFromMux(s, r)

	server := &http.Server{
		Handler:      h,
		Addr:         s.Config.ServerURL(),
		ReadTimeout:  s.Config.ReadTimeout(),
		WriteTimeout: s.Config.WriteTimeout(),
		IdleTimeout:  s.Config.IdleTimeout(),
	}

	slog.Info("Server started!", slog.String("URL", s.Config.ServerURL()))

	if err := server.ListenAndServe(); err != nil {
		slog.Error("critical startup failure", "err", err)
		os.Exit(1)
	}
}

var _ api.ServerInterface = (*Server)(nil)

// TODO: Сделать возвращение токенов
// AuthLoginUser implements [api.ServerInterface].
func (s *Server) AuthLoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusOK, user, WithType("user"))
}

// TODO: Сделать возвращение токенов
// AuthRegisterUser implements [api.ServerInterface].
func (s *Server) AuthRegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusCreated, user, WithType("user"))
}

// CoursesCreate implements [api.ServerInterface].
func (s *Server) CoursesCreate(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		course models.Course
	)

	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", WithType("error"))
		return
	}

	if err := course.CreateCourse(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error creating course", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusCreated, map[string]any{
		"course_id": course.ID,
	}, WithType("course"))
}

// CoursesCreateSection implements [api.ServerInterface].
func (s *Server) CoursesCreateSection(w http.ResponseWriter, r *http.Request, courseId int) {
	var (
		ctx    = r.Context()
		course models.Course
	)

	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", WithType("error"))
		return
	}

	if err := course.CreateCourse(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error updating course", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusCreated, map[string]any{
		"course_id": course.ID,
	}, WithType("json"))
}

// CoursesDelete implements [api.ServerInterface].
func (s *Server) CoursesDelete(w http.ResponseWriter, r *http.Request, courseId int) {
	var (
		ctx    = r.Context()
		course models.Course
	)

	if err := course.DeleteCourse(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error deleting course by id", slog.String("error", err.Error()), slog.Any("ID", courseId))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusOK, course.ID, WithType("course"))
}

// CoursesGetById implements [api.ServerInterface].
func (s *Server) CoursesGetById(w http.ResponseWriter, r *http.Request, courseId int) {
	var (
		ctx    = r.Context()
		course models.Course
	)

	if err := course.GetCourseByID(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error getting course by id", slog.String("error", err.Error()), slog.Any("ID", courseId))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusOK, course.ID, WithType("course"))
}

// CoursesGetSections implements [api.ServerInterface].
func (s *Server) CoursesGetSections(w http.ResponseWriter, r *http.Request, courseId int, params api.CoursesGetSectionsParams) {
	var (
		ctx     = r.Context()
		section models.Section
	)

	sections, err := section.GetSections(ctx, s.DB, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting sections", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusOK, sections, WithType("sections"))
}

// CoursesList implements [api.ServerInterface].
func (s *Server) CoursesList(w http.ResponseWriter, r *http.Request, params api.CoursesListParams) {
	var (
		ctx = r.Context()
		c   models.Course
	)

	courses, err := c.GetAllCourses(ctx, s.DB, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting courses", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusOK, courses, WithType("user"))
}

// CoursesUpdate implements [api.ServerInterface].
func (s *Server) CoursesUpdate(w http.ResponseWriter, r *http.Request, courseId int) {
	var (
		ctx    = r.Context()
		course models.Course
	)

	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", WithType("error"))
		return
	}

	if err := course.UpdateCourse(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error updating course", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusCreated, map[string]any{
		"course_id": course.ID,
	}, WithType("json"))
}

// EnrollmentsEnrollInCourse implements [api.ServerInterface].
func (s *Server) EnrollmentsEnrollInCourse(w http.ResponseWriter, r *http.Request, courseId int) {
	panic("unimplemented")
}

// LessonsDelete implements [api.ServerInterface].
func (s *Server) LessonsDelete(w http.ResponseWriter, r *http.Request, courseId int, sectionId int, lessonId int) {
	var (
		ctx    = r.Context()
		lesson models.Lesson
	)

	if err := lesson.DeleteLesson(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error deleting lesson", slog.String("error", err.Error()), slog.Any("ID", lessonId))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusOK, lesson.ID, WithType("lesson"))
}

// LessonsGetById implements [api.ServerInterface].
func (s *Server) LessonsGetById(w http.ResponseWriter, r *http.Request, courseId int, sectionId int, lessonId int) {
	panic("unimplemented")
}

// LessonsUpdate implements [api.ServerInterface].
func (s *Server) LessonsUpdate(w http.ResponseWriter, r *http.Request, courseId int, sectionId int, lessonId int) {
	var (
		ctx    = r.Context()
		lesson models.Lesson
	)

	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", WithType("error"))
		return
	}

	if err := lesson.UpdateLesson(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error updating lesson", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusCreated, map[string]any{
		"lesson": lesson.ID,
	}, WithType("lesson"))
}

// ProgressGetUserProgress implements [api.ServerInterface].
func (s *Server) ProgressGetUserProgress(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// ProgressUpdateLessonProgress implements [api.ServerInterface].
func (s *Server) ProgressUpdateLessonProgress(w http.ResponseWriter, r *http.Request, courseId int, lessonId int) {
	panic("unimplemented")
}

// SectionsCreateLesson implements [api.ServerInterface].
func (s *Server) SectionsCreateLesson(w http.ResponseWriter, r *http.Request, courseId int, sectionId int) {
	panic("unimplemented")
}

// SectionsDelete implements [api.ServerInterface].
func (s *Server) SectionsDelete(w http.ResponseWriter, r *http.Request, courseId int, sectionId int) {
	var (
		ctx     = r.Context()
		section models.Section
	)

	if err := section.DeleteSection(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error deleting section by id", slog.String("error", err.Error()), slog.Any("ID", sectionId))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusOK, section.ID, WithType("course"))
}

// SectionsGetById implements [api.ServerInterface].
func (s *Server) SectionsGetById(w http.ResponseWriter, r *http.Request, courseId int, sectionId int) {
	panic("unimplemented")
}

// SectionsGetLessons implements [api.ServerInterface].
func (s *Server) SectionsGetLessons(w http.ResponseWriter, r *http.Request, courseId int, sectionId int, params api.SectionsGetLessonsParams) {
	panic("unimplemented")
}

// SectionsUpdate implements [api.ServerInterface].
func (s *Server) SectionsUpdate(w http.ResponseWriter, r *http.Request, courseId int, sectionId int) {
	var (
		ctx     = r.Context()
		section models.Section
	)

	if err := json.NewDecoder(r.Body).Decode(&section); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", WithType("error"))
		return
	}

	if err := section.UpdateSection(ctx, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error updating section", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", WithType("error"))
		return
	}

	s.JSON(w, r, http.StatusCreated, map[string]any{
		"section": section.ID,
	}, WithType("section"))
}
