package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	appdb "github.com/user/u-status/internal/db"
)

type MonitorHandler struct {
	DB         *sql.DB
	NotifyCh   chan<- string // sends monitor ID to scheduler on create
	DeleteCh   chan<- string // sends monitor ID to scheduler on delete
}

type createMonitorRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
}

type updateMonitorRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
	AlertEmail      *bool  `json:"alert_email"`
	IsPublic        *bool  `json:"is_public"`
}

func (h *MonitorHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	var req createMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.URL = strings.TrimSpace(req.URL)
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if !isValidURL(req.URL) {
		respondError(w, http.StatusBadRequest, "url must start with http:// or https://")
		return
	}
	if req.IntervalSeconds == 0 {
		req.IntervalSeconds = 60
	}
	if req.IntervalSeconds < 30 {
		respondError(w, http.StatusBadRequest, "interval must be at least 30 seconds")
		return
	}

	m, err := appdb.CreateMonitor(h.DB, userID, req.Name, req.URL, req.IntervalSeconds)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	// Notify scheduler about new monitor
	if h.NotifyCh != nil {
		select {
		case h.NotifyCh <- m.ID:
		default:
		}
	}

	respondJSON(w, http.StatusCreated, m)
}

func (h *MonitorHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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

func (h *MonitorHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	existing, err := appdb.GetMonitor(h.DB, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}
	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req updateMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}
	url := strings.TrimSpace(req.URL)
	if url == "" {
		url = existing.URL
	} else if !isValidURL(url) {
		respondError(w, http.StatusBadRequest, "url must start with http:// or https://")
		return
	}
	interval := req.IntervalSeconds
	if interval == 0 {
		interval = existing.IntervalSeconds
	}
	alertEmail := existing.AlertEmail
	if req.AlertEmail != nil {
		alertEmail = *req.AlertEmail
	}
	isPublic := existing.IsPublic
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	if err := appdb.UpdateMonitor(h.DB, id, userID, name, url, interval, alertEmail, isPublic); err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	updated, _ := appdb.GetMonitor(h.DB, id)
	respondJSON(w, http.StatusOK, updated)
}

func (h *MonitorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	existing, err := appdb.GetMonitor(h.DB, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}
	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := appdb.DeleteMonitor(h.DB, id, userID); err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}

	// Notify scheduler to stop goroutine for this monitor
	if h.DeleteCh != nil {
		select {
		case h.DeleteCh <- id:
		default:
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MonitorHandler) GetChecks(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	m, err := appdb.GetMonitor(h.DB, id)
	if err != nil || m.UserID != userID {
		respondError(w, http.StatusNotFound, "monitor not found")
		return
	}

	hoursStr := r.URL.Query().Get("hours")
	hours := 24
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
			hours = h
		}
	}

	checks, err := appdb.ListChecks(h.DB, id, hours)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}
	if checks == nil {
		respondJSON(w, http.StatusOK, []any{})
		return
	}
	respondJSON(w, http.StatusOK, checks)
}

func (h *MonitorHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	m, err := appdb.GetMonitor(h.DB, id)
	if err != nil || m.UserID != userID {
		respondError(w, http.StatusNotFound, "monitor not found")
		return
	}

	stats, err := appdb.GetMonitorStats(h.DB, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func (h *MonitorHandler) GetIncidents(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	m, err := appdb.GetMonitor(h.DB, id)
	if err != nil || m.UserID != userID {
		respondError(w, http.StatusNotFound, "monitor not found")
		return
	}

	incidents, err := appdb.ListIncidents(h.DB, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server error")
		return
	}
	if incidents == nil {
		respondJSON(w, http.StatusOK, []any{})
		return
	}
	respondJSON(w, http.StatusOK, incidents)
}

func isValidURL(u string) bool {
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}
