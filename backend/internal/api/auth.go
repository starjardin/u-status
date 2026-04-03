package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/user/u-status/internal/auth"
	appdb "github.com/user/u-status/internal/db"
	"github.com/user/u-status/internal/models"
)

type AuthHandler struct {
	DB        *sql.DB
	JWTSecret string
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		respondError(w, http.StatusBadRequest, "valid email is required")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	user, err := appdb.CreateUser(h.DB, req.Email, hash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			respondError(w, http.StatusConflict, "an account with that email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	token, err := auth.IssueToken(user.ID, h.JWTSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  sanitizeUser(user),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user, err := appdb.GetUserByEmail(h.DB, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.IssueToken(user.ID, h.JWTSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  sanitizeUser(user),
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	user, err := appdb.GetUserByID(h.DB, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	respondJSON(w, http.StatusOK, sanitizeUser(user))
}

func sanitizeUser(u *models.User) map[string]any {
	return map[string]any{
		"id":         u.ID,
		"email":      u.Email,
		"plan":       u.Plan,
		"is_admin":   u.IsAdmin,
		"created_at": u.CreatedAt,
	}
}

// getUserFromDB is a package-level helper used by middleware
func getUserFromDB(db *sql.DB, userID string) (*models.User, error) {
	return appdb.GetUserByID(db, userID)
}

// getMonitorCount is a package-level helper used by middleware
func getMonitorCount(db *sql.DB, userID string) (int, error) {
	return appdb.CountMonitorsByUser(db, userID)
}
