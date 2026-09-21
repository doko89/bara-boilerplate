package admin

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"boilerplate/internal/components"
	"boilerplate/internal/domain"
	"boilerplate/internal/features/admin/templates"
	"boilerplate/internal/middleware"
	"boilerplate/internal/session"
	"boilerplate/internal/webctx"
)

type Handler struct {
	svc      *Service
	sessions *session.Manager
	log      *slog.Logger
}

func RegisterRoutes(mux *http.ServeMux, svc *Service, log *slog.Logger) {
	h := &Handler{svc: svc, log: log}
	mux.HandleFunc("GET /admin", h.overview)
	mux.HandleFunc("GET /admin/users", h.users)
	mux.HandleFunc("POST /admin/users/{id}/role", h.setRole)
	mux.HandleFunc("POST /admin/users/{id}/status", h.setStatus)
	mux.HandleFunc("POST /admin/users/{id}/delete", h.deleteUser)
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAdmin(w, r) {
		return
	}
	user := webctx.User(r.Context())
	flashType, flashMsg := webctx.ConsumeFlash(w, r)

	stats := h.svc.Overview()
	d := templates.OverviewData{
		Theme:          webctx.Theme(r),
		CSRF:           webctx.CSRFToken(r.Context()),
		FlashType:      flashType,
		FlashMsg:       flashMsg,
		TotalUsers:     stats.TotalUsers,
		ActiveUsers:    stats.ActiveUsers,
		Admins:         stats.Admins,
		NewThisWeek:    stats.NewThisWeek,
		ActiveSessions: stats.ActiveSessions,
		RecentUsers:    stats.RecentUsers,
	}

	app := components.AppPageData{
		Title:   "Admin overview",
		Active:  "overview",
		Theme:   d.Theme,
		User:    user,
		CSRF:    d.CSRF,
		Content: templates.OverviewContent(d),
	}
	components.AppLayout(app).Render(r.Context(), w)
}

func (h *Handler) users(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAdmin(w, r) {
		return
	}
	user := webctx.User(r.Context())
	flashType, flashMsg := webctx.ConsumeFlash(w, r)

	list, err := h.svc.Users()
	if err != nil {
		h.log.Error("list users failed", "error", err)
	}
	d := templates.UsersData{
		Theme:         webctx.Theme(r),
		CSRF:          webctx.CSRFToken(r.Context()),
		FlashType:     flashType,
		FlashMsg:      flashMsg,
		Users:         list,
		CurrentUserID: user.ID,
	}

	app := components.AppPageData{
		Title:   "Users",
		Active:  "users",
		Theme:   d.Theme,
		User:    user,
		CSRF:    d.CSRF,
		Content: templates.UsersContent(d),
	}
	components.AppLayout(app).Render(r.Context(), w)
}

func (h *Handler) setRole(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAdmin(w, r) {
		return
	}
	actor := webctx.User(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	role := domain.Role(r.FormValue("role"))

	switch err := h.svc.SetRole(actor, id, role); {
	case err == nil:
		webctx.SetFlash(w, "success", "Role updated.")
	case errors.Is(err, ErrSelfAction):
		webctx.SetFlash(w, "error", "You can't change your own role.")
	case errors.Is(err, ErrLastAdmin):
		webctx.SetFlash(w, "error", "The workspace needs at least one admin.")
	case errors.Is(err, ErrBadRole), errors.Is(err, ErrNotFound):
		webctx.SetFlash(w, "error", "That action isn't valid.")
	default:
		h.log.Error("set role failed", "error", err)
		webctx.SetFlash(w, "error", "Something went wrong.")
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *Handler) setStatus(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAdmin(w, r) {
		return
	}
	actor := webctx.User(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	active := r.FormValue("active") == "true"

	switch err := h.svc.SetStatus(actor, id, active); {
	case err == nil:
		if active {
			webctx.SetFlash(w, "success", "Account activated.")
		} else {
			webctx.SetFlash(w, "success", "Account deactivated and signed out everywhere.")
		}
	case errors.Is(err, ErrSelfAction):
		webctx.SetFlash(w, "error", "You can't deactivate your own account.")
	case errors.Is(err, ErrLastAdmin):
		webctx.SetFlash(w, "error", "The workspace needs at least one admin.")
	case errors.Is(err, ErrNotFound):
		webctx.SetFlash(w, "error", "That action isn't valid.")
	default:
		h.log.Error("set status failed", "error", err)
		webctx.SetFlash(w, "error", "Something went wrong.")
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAdmin(w, r) {
		return
	}
	actor := webctx.User(r.Context())
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	switch err := h.svc.Delete(actor, id); {
	case err == nil:
		webctx.SetFlash(w, "success", "User deleted.")
	case errors.Is(err, ErrSelfAction):
		webctx.SetFlash(w, "error", "You can't delete your own account.")
	case errors.Is(err, ErrLastAdmin):
		webctx.SetFlash(w, "error", "The workspace needs at least one admin.")
	case errors.Is(err, ErrNotFound):
		webctx.SetFlash(w, "error", "That action isn't valid.")
	default:
		h.log.Error("delete user failed", "error", err)
		webctx.SetFlash(w, "error", "Something went wrong.")
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func parseID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	n, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || n == 0 {
		webctx.SetFlash(w, "error", "That action isn't valid.")
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return 0, false
	}
	return uint(n), true
}
