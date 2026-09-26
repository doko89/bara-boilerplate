package settings

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"boilerplate/internal/components"
	"boilerplate/internal/features/auth"
	"boilerplate/internal/features/settings/templates"
	"boilerplate/internal/features/uploads"
	"boilerplate/internal/middleware"
	"boilerplate/internal/session"
	"boilerplate/internal/webctx"
)

type Handler struct {
	svc      *Service
	sessions *session.Manager
	avatars  *uploads.Store
}

func RegisterRoutes(mux *http.ServeMux, svc *Service, avatars *uploads.Store) {
	h := &Handler{svc: svc, avatars: avatars}
	mux.HandleFunc("GET /settings", h.page)
	mux.HandleFunc("POST /settings/profile", h.saveProfile)
	mux.HandleFunc("POST /settings/avatar", h.uploadAvatar)
	mux.HandleFunc("POST /settings/password", h.changePassword)
	mux.HandleFunc("POST /settings/sessions/revoke-others", h.revokeOthers)
}

func (h *Handler) page(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAuth(w, r) {
		return
	}
	h.render(w, r, templates.SettingsData{Errors: map[string]string{}}, http.StatusOK)
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, d templates.SettingsData, code int) {
	user := webctx.User(r.Context())
	current := webctx.Session(r.Context())
	flashType, flashMsg := webctx.ConsumeFlash(w, r)

	d.Title = "Settings | Bara"
	d.Theme = webctx.Theme(r)
	d.CSRF = webctx.CSRFToken(r.Context())
	d.FlashType = flashType
	d.FlashMsg = flashMsg
	d.Name = user.Name
	d.Email = user.Email
	d.AvatarURL = user.AvatarURL
	d.Initials = user.Initials()
	d.MaxUploadBytes = h.avatars.MaxBytes()
	d.Sessions = h.svc.Sessions(user.ID, 8)
	if current != nil {
		d.CurrentSessionID = current.ID
	}
	if d.Errors == nil {
		d.Errors = map[string]string{}
	}
	if code != http.StatusOK {
		w.WriteHeader(code)
	}

	app := components.AppPageData{
		Title:   "Settings",
		Active:  "settings",
		Theme:   d.Theme,
		User:    user,
		CSRF:    d.CSRF,
		Scripts: []string{"/static/js/avatar.js"},
		Content: templates.SettingsContent(d),
	}
	components.AppLayout(app).Render(r.Context(), w)
}

func (h *Handler) saveProfile(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAuth(w, r) {
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := auth.NormalizeEmail(r.FormValue("email"))

	errs := map[string]string{}
	switch {
	case name == "":
		errs["name"] = "Name is required."
	case len(name) < 2:
		errs["name"] = "Name must be at least 2 characters."
	case len(name) > 120:
		errs["name"] = "Name is too long."
	}
	switch {
	case email == "":
		errs["email"] = "Email is required."
	case !auth.ValidEmail(email):
		errs["email"] = "Enter a valid email address."
	}
	if len(errs) > 0 {
		h.render(w, r, templates.SettingsData{Errors: errs}, http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.UpdateProfile(webctx.User(r.Context()).ID, name, email); err != nil {
		if errors.Is(err, ErrEmailTaken) {
			h.render(w, r, templates.SettingsData{Errors: map[string]string{
				"email": "Another account already uses this email.",
			}}, http.StatusUnprocessableEntity)
			return
		}
		h.render(w, r, templates.SettingsData{Errors: map[string]string{
			"profile_form": "Something went wrong on our side. Try again.",
		}}, http.StatusInternalServerError)
		return
	}
	webctx.SetFlash(w, "success", "Profile updated.")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// uploadAvatar accepts a client-converted WebP avatar (canvas.toBlob,
// quality 0.8), stores it as /uploads/avatar/<uuid>.webp and points the
// user record at it. The previous avatar file is removed best-effort.
func (h *Handler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAuth(w, r) {
		return
	}
	user := webctx.User(r.Context())
	// Bound the body: file limit plus headroom for multipart framing.
	r.Body = http.MaxBytesReader(w, r.Body, h.avatars.MaxBytes()+(1<<20))
	f, _, err := r.FormFile("avatar")
	if err != nil {
		h.avatarError(w, r, http.StatusBadRequest, "Choose an image file first.")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		h.avatarError(w, r, http.StatusRequestEntityTooLarge, "That image is too large.")
		return
	}
	url, err := h.avatars.Save("avatar", data)
	if err != nil {
		switch {
		case errors.Is(err, uploads.ErrTooLarge):
			h.avatarError(w, r, http.StatusRequestEntityTooLarge, "That image is too large.")
		case errors.Is(err, uploads.ErrNotWebP):
			h.avatarError(w, r, http.StatusUnprocessableEntity, "Upload a WebP image converted by the form.")
		default:
			h.avatarError(w, r, http.StatusInternalServerError, "Could not save the image. Try again.")
		}
		return
	}
	if err := h.svc.UpdateAvatarURL(user.ID, url); err != nil {
		// Roll back the orphaned file; the user-visible error is the DB one.
		_ = h.avatars.Delete("avatar", strings.TrimPrefix(url, uploads.RoutePrefix+"avatar/"))
		h.avatarError(w, r, http.StatusInternalServerError, "Could not save the image. Try again.")
		return
	}
	if strings.HasPrefix(user.AvatarURL, uploads.RoutePrefix+"avatar/") {
		// Best-effort cleanup of the replaced file; a leftover is harmless
		// because its URL is no longer referenced anywhere.
		_ = h.avatars.Delete("avatar", strings.TrimPrefix(user.AvatarURL, uploads.RoutePrefix+"avatar/"))
	}
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"avatar_url": url})
		return
	}
	webctx.SetFlash(w, "success", "Profile photo updated.")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// avatarError answers JSON for the drag-and-drop fetch flow and falls back
// to a flash + redirect for plain form posts.
func (h *Handler) avatarError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return
	}
	webctx.SetFlash(w, "error", msg)
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAuth(w, r) {
		return
	}
	user := webctx.User(r.Context())
	current := webctx.Session(r.Context())

	currentPw := r.FormValue("current_password")
	newPw := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	errs := map[string]string{}
	switch {
	case currentPw == "":
		errs["current_password"] = "Enter your current password."
	case newPw == "":
		errs["password"] = "New password is required."
	case !auth.ValidPassword(newPw):
		errs["password"] = "Use at least 8 characters."
	}
	if confirm != newPw {
		errs["confirm_password"] = "Passwords don't match."
	}
	if len(errs) > 0 {
		h.render(w, r, templates.SettingsData{Errors: errs}, http.StatusUnprocessableEntity)
		return
	}

	sessionID := ""
	if current != nil {
		sessionID = current.ID
	}
	if err := h.svc.ChangePassword(user, sessionID, currentPw, newPw); err != nil {
		if errors.Is(err, ErrWrongPassword) {
			h.render(w, r, templates.SettingsData{Errors: map[string]string{
				"current_password": "Current password is incorrect.",
			}}, http.StatusUnprocessableEntity)
			return
		}
		h.render(w, r, templates.SettingsData{Errors: map[string]string{
			"password_form": "Something went wrong on our side. Try again.",
		}}, http.StatusInternalServerError)
		return
	}
	webctx.SetFlash(w, "success", "Password updated. Other sessions were signed out.")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (h *Handler) revokeOthers(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAuth(w, r) {
		return
	}
	user := webctx.User(r.Context())
	current := webctx.Session(r.Context())
	sessionID := ""
	if current != nil {
		sessionID = current.ID
	}
	h.svc.RevokeOtherSessions(user.ID, sessionID)
	webctx.SetFlash(w, "success", "All other sessions were signed out.")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
