package handlers

import (
	"context"
	"encoding/json"
	"handbooks/internal/api"
	"handbooks/internal/config"
	"handbooks/internal/models"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type Server struct {
	DB     *pgx.Conn
	Config *config.Config
	Logger *slog.Logger
}

func NewServer(db *pgx.Conn, config *config.Config, logger *slog.Logger) *Server {
	return &Server{
		DB:     db,
		Config: config,
		Logger: logger,
	}
}

var _ api.ServerInterface = (*Server)(nil)

// GetCourses implements [api.ServerInterface].
func (s *Server) GetCourses(w http.ResponseWriter, r *http.Request, params api.GetCoursesParams) {
	s.Logger.Info("Hello")
	s.httpResponse(w, http.StatusOK, map[string]interface{}{
		"Hello": "World",
	}, "json")
}

// GetCoursesCourseId implements [api.ServerInterface].
func (s *Server) GetCoursesCourseId(w http.ResponseWriter, r *http.Request, courseId int) {
	panic("unimplemented")
}

// GetCoursesCourseIdSections implements [api.ServerInterface].
func (s *Server) GetCoursesCourseIdSections(w http.ResponseWriter, r *http.Request, courseId int) {
	panic("unimplemented")
}

// GetMeProgress implements [api.ServerInterface].
func (s *Server) GetMeProgress(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// PatchMeCoursesCourseIdLessonsLessonIdProgress implements [api.ServerInterface].
func (s *Server) PatchMeCoursesCourseIdLessonsLessonIdProgress(w http.ResponseWriter, r *http.Request, courseId int, lessonId int) {
	panic("unimplemented")
}

// PostAuthLogin implements [api.ServerInterface].
func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.Logger.Error("Error decoding request body", slog.String("error", err.Error()))
		s.httpResponse(w, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	var u models.User
	row := s.DB.QueryRow(context.Background(), "SELECT * FROM users WHERE email = $1 and password_hash = $2", user.Email, user.PasswordHash)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarURL, &u.Role, &u.CreatedAt, &u.LastLoginAt); err != nil {
		s.Logger.Error("Error scanning database", slog.String("error", err.Error()))
		s.httpResponse(w, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.httpResponse(w, http.StatusOK, map[string]any{
		"user": user,
	}, "json")
}

// PostAuthRegister implements [api.ServerInterface].
func (s *Server) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.Logger.Error("Error decoding request body", slog.String("error", err.Error()))
		s.httpResponse(w, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	_, err := s.DB.Exec(context.Background(), "INSERT INTO users (email, password_hash, full_name, avatar_url, role, created_at, last_login_at) VALUES ($1, $2, $3, $4, $5, $6, $7)", user.Email, user.PasswordHash, user.FullName, user.AvatarURL, "student", time.Now(), time.Now())
	if err != nil {
		s.Logger.Error("Error inserting user", slog.String("error", err.Error()))
		s.httpResponse(w, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.httpResponse(w, http.StatusCreated, map[string]any{
		"user": user,
	}, "json")
}

// PostCourses implements [api.ServerInterface].
func (s *Server) PostCourses(w http.ResponseWriter, r *http.Request) {
	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		s.Logger.Error("Error decoding request body", slog.String("error", err.Error()))
		s.httpResponse(w, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	var user models.User
	if _, err := s.DB.QueryRow(context.Background(), "SELECT * FROM users WHERE id = $1", course.CreatedByID).Scan(&user.ID, &user.Email, &user.Name, &user.Role); err != nil {
		s.Logger.Error("Error fetching user", slog.String("error", err.Error()))
		s.httpResponse(w, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	_, err := s.DB.Exec(context.Background(), "INSERT INTO courses (name, description, instructor_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", course.Name, course.Description, course.InstructorID, time.Now(), time.Now())
	if err != nil {
		s.Logger.Error("Error inserting course", slog.String("error", err.Error()))
		s.httpResponse(w, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.httpResponse(w, http.StatusCreated, map[string]any{
		"user": user,
	}, "json")
}

// PostCoursesCourseIdEnroll implements [api.ServerInterface].
func (s *Server) PostCoursesCourseIdEnroll(w http.ResponseWriter, r *http.Request, courseId int) {
	panic("unimplemented")
}

func (s *Server) httpResponse(w http.ResponseWriter, status int, data any, respType string) {
	resp := map[string]any{
		"status": status,
		"type":   respType,
		"data":   data,
	}

	respData, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		s.Logger.Error("Error marshaling response", slog.String("error", marshalErr.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(respData); err != nil {
		s.Logger.Error("Error writing response", slog.String("error", err.Error()))
	}
}
