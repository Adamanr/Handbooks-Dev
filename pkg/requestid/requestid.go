package requestid

import (
	"context"
	"net/http"

	"github.com/rs/xid"
)

type ctxKey string

const requestIDKey ctxKey = "request-id"

// MiddlewareRequestID — добавляет request-id в контекст и (опционально) в заголовки
func MiddlewareRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = xid.New().String()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, rid)

		w.Header().Set("X-Request-ID", rid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
