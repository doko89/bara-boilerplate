package landing

import (
	"net/http"

	"boilerplate/internal/features/landing/templates"
	"boilerplate/internal/webctx"
)

// RegisterRoutes mounts the public landing page.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		templates.View(templates.LandingData{
			Title: "Bara: production-ready Go boilerplate",
			Theme: webctx.Theme(r),
			User:  webctx.User(r.Context()),
		}).Render(r.Context(), w)
	})
}
