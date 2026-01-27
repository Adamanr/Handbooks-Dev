package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"handbooks/internal/api"
	"handbooks/internal/models"
	storage "handbooks/pkg/storage"

	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/huandu/go-sqlbuilder"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	ID      uuid.UUID `json:"id"`
	Email   string    `json:"email"`
	Role    string    `json:"role"`
	TokenID string    `json:"token_id"`
	jwt.RegisteredClaims
}

// AuthLoginUser — вход
func (s *Server) AuthLoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req api.AuthLoginUserJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.JSON(w, r, http.StatusBadRequest, "ошибка при получении данных", "error")
		return
	}

	user, err := storage.GetOne[models.User](ctx, s.DB, "users", func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("email", req.Email))
	})
	if err != nil {
		slog.WarnContext(ctx, "user not found or db error", "email", req.Email, "err", err)
		s.JSON(w, r, http.StatusUnauthorized, "пользователь по email не найден", "error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.JSON(w, r, http.StatusUnauthorized, "Неверный пароль", "error")
		return
	}

	s.issueTokens(w, r, user)
}

// AuthRegisterUser — регистрация + сразу логин (токены)
func (s *Server) AuthRegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req api.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.JSON(w, r, http.StatusInternalServerError, "Server error", "error")
		return
	}

	now := time.Now()
	uuid := uuid.New()
	user := models.User{
		ID:           uuid,
		Slug:         GenerateUserSlug(req.FullName, uuid),
		CreatedAt:    now,
		UpdatedAt:    now,
		Email:        string(req.Email),
		PasswordHash: string(passwordHash),
		FullName:     req.FullName,
	}

	if err := storage.Create(ctx, "users", user, s.DB); err != nil {
		slog.ErrorContext(ctx, "Error creating users", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.issueTokens(w, r, &user)
}

// AuthRefreshToken implements [api.ServerInterface].
func (s *Server) AuthRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		s.JSON(w, r, http.StatusUnauthorized, "Missing refresh token", "error")
		return
	}
	refreshStr := cookie.Value
	if refreshStr == "" {
		s.JSON(w, r, http.StatusUnauthorized, "Empty refresh token", "error")
		return
	}
	claimsValue := ctx.Value("user")

	claims, ok := claimsValue.(*Claims)
	if !ok {
		slog.ErrorContext(ctx, "Error parsing claims", slog.Any("Claims", claimsValue))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}

	userID := claims.ID
	if userID == uuid.Nil {
		s.JSON(w, r, http.StatusUnauthorized, "Missing user ID in token", "error")
		return
	}

	key := "refresh_hash:" + refreshStr
	if _, err := s.Redis.Get(ctx, key).Result(); err == redis.Nil {
		s.JSON(w, r, http.StatusUnauthorized, "No active refresh token", "error")
		return
	}

	if err != nil {
		slog.ErrorContext(ctx, "redis get error", "err", err)
		s.JSON(w, r, http.StatusInternalServerError, "Server error", "error")
		return
	}

	user, err := storage.GetOne[models.User](ctx, s.DB, "users", func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", userID))
	})
	if err != nil {
		slog.ErrorContext(ctx, "user not found for refresh", "user_id", userID, "err", err)
		s.JSON(w, r, http.StatusUnauthorized, "User not found", "error")
		return
	}

	newAccess, err := s.generateAccessToken(user, s.Config.RedisAccessTokenDur())
	if err != nil {
		s.JSON(w, r, http.StatusInternalServerError, "Token generation error", "error")
		return
	}

	newRefresh, err := s.generateRefreshToken()
	if err != nil {
		s.JSON(w, r, http.StatusInternalServerError, "Token generation error", "error")
		return
	}

	if err := s.Redis.Del(ctx, key).Err(); err != nil {
		slog.ErrorContext(ctx, "redis del old refresh failed", "err", err)
	}

	if err := s.Redis.Set(ctx, key, "valid", s.Config.RedisRefreshTokenDur()).Err(); err != nil {
		slog.ErrorContext(ctx, "redis set new refresh failed", "err", err)
		s.JSON(w, r, http.StatusInternalServerError, "Server error", "error")
		return
	}

	access := "access_hash:" + newAccess
	if err := s.Redis.Set(ctx, access, "valid", s.Config.RedisAccessTokenDur()).Err(); err != nil {
		slog.ErrorContext(ctx, "redis set new access failed", "err", err)
		s.JSON(w, r, http.StatusInternalServerError, "Server error", "error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefresh,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(s.Config.RedisRefreshTokenDur()),
	})

	s.JSON(w, r, http.StatusOK, map[string]any{
		"access_token": newAccess,
		"expires_in":   86400, // 24 часа
	}, "auth")
}

// UsersDeleteById implements [api.ServerInterface].
func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request, userID string) {
	var ctx = r.Context()

	if err := storage.Delete[models.User](ctx, "users", s.DB, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", userID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error deleting user by id", slog.String("error", err.Error()), slog.Any("ID", userID))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, true, "user")
}

// UsersDeleteCurrent implements [api.ServerInterface].
func (s *Server) DeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
	)
	claimsValue := ctx.Value("user")

	claims, ok := claimsValue.(*Claims)
	if !ok {
		slog.ErrorContext(ctx, "Error parsing claims", slog.Any("Claims", claimsValue))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}
	userID := claims.ID
	if err := storage.Delete[models.User](ctx, "users", s.DB, func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", userID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error deleting user by ID", slog.String("error", err.Error()), slog.Any("ID", userID))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusOK, true, "user")
}

// UsersGetCurrent implements [api.ServerInterface].
func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
	)
	claimsValue := ctx.Value("user")

	claims, ok := claimsValue.(*Claims)
	if !ok {
		slog.ErrorContext(ctx, "Error parsing claims", slog.Any("Claims", claimsValue))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}

	userID := claims.ID

	user, err := storage.GetOne[models.User](ctx, s.DB, "users", func(sb *sqlbuilder.SelectBuilder) {
		sb.Where(sb.Equal("id", userID))
	})

	if err != nil {
		slog.ErrorContext(ctx, "Error getting user by ID",
			slog.Any("ID", userID),
			slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}

	s.JSON(w, r, http.StatusOK, user, "user")
}

// UpdateCurrentUser implements [api.ServerInterface].
func (s *Server) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = r.Context()
		user models.User
	)

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		slog.ErrorContext(ctx, "Error decoding request body", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusBadRequest, "Invalid request body", "error")
		return
	}

	claimsValue := ctx.Value("user")

	claims, ok := claimsValue.(*Claims)
	if !ok {
		slog.ErrorContext(ctx, "Error parsing claims", slog.Any("Claims", claimsValue))
		s.JSON(w, r, http.StatusInternalServerError, nil, "internal server error")
		return
	}

	userID := claims.ID
	if err := storage.Update[models.User](ctx, "users", user, s.DB, func(sb *sqlbuilder.UpdateBuilder) {
		sb.Where(sb.Equal("id", userID))
	}); err != nil {
		slog.ErrorContext(ctx, "Error updating user", slog.String("error", err.Error()))
		s.JSON(w, r, http.StatusInternalServerError, "Internal server error", "error")
		return
	}

	s.JSON(w, r, http.StatusCreated, user, "user")
}

// GetUserProgress implements [api.ServerInterface].
func (s *Server) GetUserProgress(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// UpdateLessonProgress implements [api.ServerInterface].
func (s *Server) UpdateLessonProgress(w http.ResponseWriter, r *http.Request, courseID string, lessonID string) {
	panic("unimplemented")
}

// issueTokens — общая функция выдачи токенов (логин + регистрация)
func (s *Server) issueTokens(w http.ResponseWriter, r *http.Request, user *models.User) {
	access, err := s.generateAccessToken(user, s.Config.RedisAccessTokenDur())
	if err != nil {
		slog.ErrorContext(r.Context(), "generate access failed", "err", err)
		s.JSON(w, r, http.StatusInternalServerError, "Token generation error", "error")
		return
	}

	refresh, err := s.generateRefreshToken()
	if err != nil {
		slog.ErrorContext(r.Context(), "generate refresh failed", "err", err)
		s.JSON(w, r, http.StatusInternalServerError, "Token generation error", "error")
		return
	}

	key := fmt.Sprintf("refresh_hash:%v", refresh)
	err = s.Redis.Set(r.Context(), key, "valid", s.Config.RedisAccessTokenDur()).Err()
	if err != nil {
		slog.ErrorContext(r.Context(), "redis set refresh failed", "err", err)
		s.JSON(w, r, http.StatusInternalServerError, "Server error", "error")
		return
	}

	accessKey := fmt.Sprintf("access_hash:%v", access)
	err = s.Redis.Set(r.Context(), accessKey, "valid", s.Config.RedisAccessTokenDur()).Err()
	if err != nil {
		slog.ErrorContext(r.Context(), "redis set access failed", "err", err)
		s.JSON(w, r, http.StatusInternalServerError, "Server error", "error")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
	})

	s.JSON(w, r, http.StatusOK, map[string]any{
		"access_token": access,
		"expires_in":   86400, // 24 часа
	}, "auth")
}

// Вспомогательные функции

func (s *Server) generateAccessToken(user *models.User, duration time.Duration) (string, error) {
	return s.generateJWT(user, duration)
}

func (s *Server) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Server) generateJWT(user *models.User, lifetime time.Duration) (string, error) {
	if len(s.JwtKey) == 0 {
		return "", errors.New("jwt key not set")
	}

	tokenID := hex.EncodeToString([]byte(time.Now().String() + user.ID.String()))

	claims := Claims{
		ID:      user.ID,
		Email:   user.Email,
		Role:    user.Role,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
			Issuer:    s.Config.JwtOpt.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JwtKey)
}

func (s *Server) deleteRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "refresh_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// ProgressGetUserProgress implements [api.ServerInterface].
func (s *Server) ProgressGetUserProgress(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
func GenerateUserSlug(username string, userID uuid.UUID) string {
	if username == "" {
		username = "user"
	}

	base := slug.Make(username)

	shortUUID := strings.ReplaceAll(userID.String(), "-", "")[:8]

	return base + "-" + shortUUID
}
