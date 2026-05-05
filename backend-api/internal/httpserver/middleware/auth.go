package middleware

import (
	"context"
	"demo-server/internal/app"
	"demo-server/internal/httpserver/models"
	"demo-server/internal/storage"
	"encoding/json"
	"net/http"
	"strings"
)

type adminContextKey struct{}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := authToken(r.Header.Get("Authorization"))
			if token == "" {
				writeJSON(w, http.StatusUnauthorized, models.Response{
					IsOk:    false,
					Message: "unauthorized",
				})
				return
			}

			admin, err := app.ValidateJwt(token, jwtSecret)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, models.Response{
					IsOk:    false,
					Message: "unauthorized",
				})
				return
			}

			ctx := context.WithValue(r.Context(), adminContextKey{}, admin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminFromContext(ctx context.Context) (*storage.Admin, bool) {
	admin, ok := ctx.Value(adminContextKey{}).(*storage.Admin)
	return admin, ok
}

func authToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	if len(parts) == 1 {
		return parts[0]
	}

	return ""
}

func writeJSON(w http.ResponseWriter, status int, response models.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
