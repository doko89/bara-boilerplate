package dashboard

import (
	"net/http"
	"time"

	"boilerplate/internal/components"
	"boilerplate/internal/features/dashboard/templates"
	"boilerplate/internal/middleware"
	"boilerplate/internal/webctx"
)

type Handler struct {
	svc *Service
}

func RegisterRoutes(mux *http.ServeMux, svc *Service) {
	h := &Handler{svc: svc}
	mux.HandleFunc("GET /dashboard", h.overview)
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAuth(w, r) {
		return
	}
	user := webctx.User(r.Context())

	flashType, flashMsg := webctx.ConsumeFlash(w, r)
	stats := h.svc.Overview(user)

	lastSignIn := "Never"
	if stats.HasSignIn {
		lastSignIn = components.TimeAgo(stats.LastSignIn)
	}

	d := templates.OverviewData{
		Theme:       webctx.Theme(r),
		CSRF:        webctx.CSRFToken(r.Context()),
		FlashType:   flashType,
		FlashMsg:    flashMsg,
		Greeting:    components.Greeting(time.Now()) + ", " + components.FirstName(user.Name) + ".",
		SignInCount: stats.SignInCount,
		LastSignIn:  lastSignIn,
		MemberSince: stats.MemberSince,
		Recent:      stats.Recent,
	}

	app := components.AppPageData{
		Title:   "Overview",
		Active:  "overview",
		Theme:   d.Theme,
		User:    user,
		CSRF:    d.CSRF,
		Content: templates.OverviewContent(d),
	}
	components.AppLayout(app).Render(r.Context(), w)
}
