package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	appdb "github.com/user/u-status/internal/db"
)

type AdminHandler struct {
	DB *sql.DB
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := appdb.ListAllUsers(h.DB)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	result := make([]map[string]any, 0, len(users))
	for _, u := range users {
		result = append(result, sanitizeUser(u))
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *AdminHandler) GetUserMonitors(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user id is required")
		return
	}

	monitors, err := appdb.ListMonitorsByUser(h.DB, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}
	if monitors == nil {
		respondJSON(w, http.StatusOK, []any{})
		return
	}
	respondJSON(w, http.StatusOK, monitors)
}
