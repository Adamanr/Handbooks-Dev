package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// APIResponse — единая оболочка для всех JSON-ответов
type APIResponse struct {
	Status    int    `json:"status"`         // HTTP-подобный код
	Success   bool   `json:"success"`        // true если 2xx
	Type      string `json:"type,omitempty"` // "user", "auth", "error", etc
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`      // человекочитаемое сообщение
	ErrorCode string `json:"error_code,omitempty"` // машина-читаемый код ошибки
	RequestID string `json:"request_id,omitempty"` // для трассировки
}

// JSON - форматирует ответ в JSON и отправляет его клиенту
func (s *Server) JSON(w http.ResponseWriter, r *http.Request, status int, payload any, opts ...ResponseOption) {
	options := responseOptions{
		respType:  "",
		requestID: extractRequestID(r),
	}

	for _, opt := range opts {
		opt(&options)
	}

	resp := APIResponse{
		Status:    status,
		Success:   status >= 200 && status < 300,
		Type:      options.respType,
		RequestID: options.requestID,
	}

	switch v := payload.(type) {
	case nil:
		// ничего не кладём
	case error:
		resp.Error = v.Error()
		if e, ok := v.(interface{ Code() string }); ok {
			resp.ErrorCode = e.Code()
		}
	default:
		resp.Data = v
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("json encode failed after headers written",
			slog.Int("status", status),
			slog.String("error", err.Error()),
			slog.String("request_id", options.requestID),
		)
	}
}

type ctxKey string

const requestIDKey ctxKey = "request-id"

// extractRequestID - извлекает идентификатор запроса из контекста или заголовка X-Request-ID
func extractRequestID(r *http.Request) string {
	if r == nil {
		return ""
	}

	if v := r.Context().Value(requestIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}

	return ""
}

// ResponseOption - опции для форматирования ответа
type ResponseOption func(*responseOptions)

// WithType - устанавливает тип ответа
func WithType(t string) ResponseOption {
	return func(o *responseOptions) { o.respType = t }
}

// WithRequestID - устанавливает идентификатор запроса
func WithRequestID(id string) ResponseOption {
	return func(o *responseOptions) { o.requestID = id }
}

// responseOptions - опции для форматирования ответа
type responseOptions struct {
	respType  string
	requestID string
}
