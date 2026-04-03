package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/u-status/internal/auth"
)

type contextKey string

const ctxUserID contextKey = "userID"

func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

func JWTMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				respondError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ParseToken(tokenStr, jwtSecret)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnlyMiddleware(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			user, err := getUserFromDB(database, userID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "could not fetch user")
				return
			}
			if !user.IsAdmin {
				respondError(w, http.StatusForbidden, "admin access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func PlanLimitsMiddleware(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/monitors") {
				next.ServeHTTP(w, r)
				return
			}

			userID := GetUserID(r.Context())
			if userID == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := getUserFromDB(database, userID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "could not fetch user")
				return
			}

			limit := getPlanLimit(string(user.Plan))
			if limit >= 0 {
				count, err := getMonitorCount(database, userID)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "could not check monitor count")
					return
				}
				if count >= limit {
					respondError(w, http.StatusForbidden, "upgrade your plan to add more monitors")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getPlanLimit(plan string) int {
	switch plan {
	case "pro":
		return 20
	case "agency":
		return -1
	default:
		return 3
	}
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
