package handlers

import (
	"encoding/json"
	"handbooks/internal/models"
	storage "handbooks/pkg/storage"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/huandu/go-sqlbuilder"
)

// CoursesGetSections implements [api.ServerInterface].
func (s *Server) GetSections(w http.ResponseWriter, r *http.Request, courseID string) {
	var ctx = r.Context()

	sections, err := storage.GetAll[models.Section](ctx, "sections", s.DB)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting sections", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, sections, "sections")
}

// CreateSection implements [api.ServerInterface].
func (s *Server) CreateSection(w http.ResponseWriter, r *http.Request, courseID string) {
	var (
		ctx     = r.Context()
		section models.Section
	)

	if err := json.NewDecoder(r.Body).Decode(&section); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	time := time.Now()
	section.ID = uuid.New()
	section.Slug = slug.Make(section.Title)
	section.CreatedID = ctx.Value("user").(*Claims).ID
	section.CreatedAt = time
	section.UpdatedAt = time

	if err := storage.Create(ctx, "sections", section, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error creating section", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusCreated, section, "section")
}

// DeleteSection implements [api.ServerInterface].
func (s *Server) DeleteSection(w http.ResponseWriter, r *http.Request, courseID string, sectionID string) {
	var ctx = r.Context()

	if err := storage.Delete[models.Section](ctx, "sections", s.DB, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", sectionID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error deleting section by ID", slog.String("error", err.Error()), slog.Any("ID", sectionID))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, true, "section")
}

// GetSectionByID implements [api.ServerInterface].
func (s *Server) GetSectionByID(w http.ResponseWriter, r *http.Request, courseID string, sectionID string) {
	ctx := r.Context()

	section, err := storage.GetOne[models.Section](ctx, s.DB, "sections", func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", sectionID))
	})

	if err != nil {
		slog.ErrorContext(ctx, "Error getting section by id",
			slog.String("id", sectionID),
			slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}

	s.JSON(w, r, http.StatusOK, section, "section")
}

// UpdateSection implements [api.ServerInterface].
func (s *Server) UpdateSection(w http.ResponseWriter, r *http.Request, courseID string, sectionID string) {
	var (
		ctx     = r.Context()
		section models.Section
	)

	if err := json.NewDecoder(r.Body).Decode(&section); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	if err := storage.Update[models.Section](ctx, "sections", section, s.DB, func(sb *sqlbuilder.UpdateBuilder) {
		sb.Where(sb.Equal("id", sectionID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error updating section", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusCreated, section, "section")
}
