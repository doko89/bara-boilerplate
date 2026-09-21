// Package server wires config, database, middleware and every feature into a
// single http.Handler.
package server

import (
	"log/slog"
	"net/http"

	"boilerplate/internal/components"
	"boilerplate/internal/config"
	"boilerplate/internal/features/admin"
	"boilerplate/internal/features/auth"
	"boilerplate/internal/features/dashboard"
	"boilerplate/internal/features/landing"
	"boilerplate/internal/features/settings"
	"boilerplate/internal/middleware"
	"boilerplate/internal/session"
	"boilerplate/internal/webctx"

	"gorm.io/gorm"
)

func New(cfg config.Config, db *gorm.DB, log *slog.Logger) http.Handler {
	manager := session.NewManager(db, cfg.SessionTTL)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", staticHandler())

	landing.RegisterRoutes(mux)

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, manager, log)
	auth.RegisterRoutes(mux, authSvc, manager, cfg, log)

	dashSvc := dashboard.NewService(manager)
	dashboard.RegisterRoutes(mux, dashSvc)

	settingsRepo := settings.NewRepository(db)
	settingsSvc := settings.NewService(settingsRepo, manager)
	settings.RegisterRoutes(mux, settingsSvc)

	adminRepo := admin.NewRepository(db)
	adminSvc := admin.NewService(adminRepo, manager)
	admin.RegisterRoutes(mux, adminSvc, log)

	root := withNotFound(mux)

	return middleware.Chain(root,
		middleware.Recover(log, renderServerError),
		middleware.RequestLog(log),
		middleware.SecurityHeaders,
		middleware.LoadSession(manager),
		middleware.NewCSRF(cfg.SecureCookies),
	)
}

// withNotFound renders the styled 404 page whenever no route matches.
func withNotFound(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := mux.Handler(r)
		if pattern == "" {
			w.WriteHeader(http.StatusNotFound)
			components.ErrorPage(
				http.StatusNotFound,
				"Page not found",
				"The page you're looking for doesn't exist or was moved.",
				webctx.Theme(r),
				webctx.User(r.Context()),
			).Render(r.Context(), w)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func renderServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	components.ErrorPage(
		http.StatusInternalServerError,
		"Something broke",
		"An unexpected error occurred. The details are in the server logs.",
		webctx.Theme(r),
		webctx.User(r.Context()),
	).Render(r.Context(), w)
}

func staticHandler() http.Handler {
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir("web/static")))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static" || r.URL.Path == "/static/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fs.ServeHTTP(w, r)
	})
}
