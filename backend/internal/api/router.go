package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(db *sql.DB, jwtSecret string, allowedOrigins []string, notifyCh chan<- string, deleteCh chan<- string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	authHandler := &AuthHandler{DB: db, JWTSecret: jwtSecret}
	monitorHandler := &MonitorHandler{DB: db, NotifyCh: notifyCh, DeleteCh: deleteCh}

	// Public routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(JWTMiddleware(jwtSecret))

		r.Get("/api/me", authHandler.Me)

		r.Route("/api/monitors", func(r chi.Router) {
			r.Use(PlanLimitsMiddleware(db))
			r.Post("/", monitorHandler.Create)
			r.Get("/", monitorHandler.List)
			r.Put("/{id}", monitorHandler.Update)
			r.Delete("/{id}", monitorHandler.Delete)
			r.Get("/{id}/checks", monitorHandler.GetChecks)
			r.Get("/{id}/stats", monitorHandler.GetStats)
			r.Get("/{id}/incidents", monitorHandler.GetIncidents)
		})
	})

	return r
}
