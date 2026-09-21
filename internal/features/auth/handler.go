package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"boilerplate/internal/components"
	"boilerplate/internal/config"
	"boilerplate/internal/domain"
	"boilerplate/internal/features/auth/templates"
	"boilerplate/internal/middleware"
	"boilerplate/internal/session"
	"boilerplate/internal/webctx"

	"github.com/a-h/templ"
)

type Handler struct {
	svc      *Service
	sessions *session.Manager
	cfg      config.Config
	log      *slog.Logger
}

func RegisterRoutes(mux *http.ServeMux, svc *Service, sessions *session.Manager, cfg config.Config, log *slog.Logger) {
	h := &Handler{svc: svc, sessions: sessions, cfg: cfg, log: log}
	mux.HandleFunc("GET /login", h.loginForm)
	mux.HandleFunc("POST /login", h.loginSubmit)
	mux.HandleFunc("GET /register", h.registerForm)
	mux.HandleFunc("POST /register", h.registerSubmit)
	mux.HandleFunc("GET /forgot-password", h.forgotForm)
	mux.HandleFunc("POST /forgot-password", h.forgotSubmit)
	mux.HandleFunc("GET /reset-password", h.resetForm)
	mux.HandleFunc("POST /reset-password", h.resetSubmit)
	mux.HandleFunc("POST /logout", h.logout)
}

func (h *Handler) authData(r *http.Request, panelKey string, content templ.Component) components.AuthPageData {
	panelTitle, panelText := components.AuthPanelCopy(panelKey)
	return components.AuthPageData{
		Title:      "",
		Theme:      webctx.Theme(r),
		User:       webctx.User(r.Context()),
		CSRF:       webctx.CSRFToken(r.Context()),
		PanelTitle: panelTitle,
		PanelText:  panelText,
		Content:    content,
	}
}

func (h *Handler) renderLogin(w http.ResponseWriter, r *http.Request, d templates.LoginData, code int) {
	d.Title = "Sign in | Bara"
	d.Theme = webctx.Theme(r)
	d.CSRF = webctx.CSRFToken(r.Context())
	if d.Errors == nil {
		d.Errors = map[string]string{}
	}
	d.DemoHint = h.cfg.IsDev()
	d.DemoAdmin = h.cfg.SeedAdminEmail + " / " + h.cfg.SeedAdminPassword
	d.DemoUser = "aulia@bara.dev / User1234"
	if code != http.StatusOK {
		w.WriteHeader(code)
	}
	components.AuthLayout(h.authData(r, "login", templates.LoginCard(d))).Render(r.Context(), w)
}

func (h *Handler) loginForm(w http.ResponseWriter, r *http.Request) {
	if middleware.RedirectIfAuthenticated(w, r) {
		return
	}
	flashType, flashMsg := webctx.ConsumeFlash(w, r)
	h.renderLogin(w, r, templates.LoginData{
		Email:     "",
		Errors:    map[string]string{},
		FlashType: flashType,
		FlashMsg:  flashMsg,
	}, http.StatusOK)
}

func (h *Handler) loginSubmit(w http.ResponseWriter, r *http.Request) {
	email := NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	errs := map[string]string{}
	if email == "" {
		errs["email"] = "Email is required."
	} else if !ValidEmail(email) {
		errs["email"] = "Enter a valid email address."
	}
	if password == "" {
		errs["password"] = "Password is required."
	}
	if len(errs) > 0 {
		h.renderLogin(w, r, templates.LoginData{Email: email, Errors: errs}, http.StatusUnprocessableEntity)
		return
	}

	user, err := h.svc.Login(email, password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			h.renderLogin(w, r, templates.LoginData{Email: email, Errors: map[string]string{
				"form": "Invalid email or password.",
			}}, http.StatusUnauthorized)
		case errors.Is(err, ErrAccountDisabled):
			h.renderLogin(w, r, templates.LoginData{Email: email, Errors: map[string]string{
				"form": "This account has been deactivated.",
			}}, http.StatusForbidden)
		default:
			h.log.Error("login failed", "error", err)
			h.renderLogin(w, r, templates.LoginData{Email: email, Errors: map[string]string{
				"form": "Something went wrong on our side. Try again.",
			}}, http.StatusInternalServerError)
		}
		return
	}

	h.startSession(w, r, user, "Welcome back, "+components.FirstName(user.Name)+".")
}

func (h *Handler) registerForm(w http.ResponseWriter, r *http.Request) {
	if middleware.RedirectIfAuthenticated(w, r) {
		return
	}
	flashType, flashMsg := webctx.ConsumeFlash(w, r)
	h.renderRegister(w, r, templates.RegisterData{
		Errors:    map[string]string{},
		FlashType: flashType,
		FlashMsg:  flashMsg,
	}, http.StatusOK)
}

func (h *Handler) registerSubmit(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	email := NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

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
	case !ValidEmail(email):
		errs["email"] = "Enter a valid email address."
	}
	switch {
	case password == "":
		errs["password"] = "Password is required."
	case !ValidPassword(password):
		errs["password"] = "Use at least 8 characters."
	}
	if confirm != password {
		errs["confirm_password"] = "Passwords don't match."
	}
	if len(errs) > 0 {
		h.renderRegister(w, r, templates.RegisterData{Name: name, Email: email, Errors: errs}, http.StatusUnprocessableEntity)
		return
	}

	user, err := h.svc.Register(name, email, password)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			h.renderRegister(w, r, templates.RegisterData{Name: name, Email: email, Errors: map[string]string{
				"email": "An account with this email already exists.",
			}}, http.StatusUnprocessableEntity)
			return
		}
		h.log.Error("register failed", "error", err)
		h.renderRegister(w, r, templates.RegisterData{Name: name, Email: email, Errors: map[string]string{
			"form": "Something went wrong on our side. Try again.",
		}}, http.StatusInternalServerError)
		return
	}

	h.startSession(w, r, user, "Your account is ready. Welcome, "+components.FirstName(user.Name)+".")
}

func (h *Handler) renderRegister(w http.ResponseWriter, r *http.Request, d templates.RegisterData, code int) {
	d.Title = "Create account | Bara"
	d.Theme = webctx.Theme(r)
	d.CSRF = webctx.CSRFToken(r.Context())
	if d.Errors == nil {
		d.Errors = map[string]string{}
	}
	if code != http.StatusOK {
		w.WriteHeader(code)
	}
	components.AuthLayout(h.authData(r, "register", templates.RegisterCard(d))).Render(r.Context(), w)
}

func (h *Handler) forgotForm(w http.ResponseWriter, r *http.Request) {
	if middleware.RedirectIfAuthenticated(w, r) {
		return
	}
	flashType, flashMsg := webctx.ConsumeFlash(w, r)
	h.renderForgot(w, r, templates.ForgotData{
		Errors:    map[string]string{},
		FlashType: flashType,
		FlashMsg:  flashMsg,
	}, http.StatusOK)
}

func (h *Handler) forgotSubmit(w http.ResponseWriter, r *http.Request) {
	email := NormalizeEmail(r.FormValue("email"))

	errs := map[string]string{}
	if email == "" {
		errs["email"] = "Email is required."
	} else if !ValidEmail(email) {
		errs["email"] = "Enter a valid email address."
	}
	if len(errs) > 0 {
		h.renderForgot(w, r, templates.ForgotData{Email: email, Errors: errs}, http.StatusUnprocessableEntity)
		return
	}

	token, err := h.svc.RequestPasswordReset(email)
	if err != nil {
		h.log.Error("password reset request failed", "error", err)
	}
	if token != "" {
		// No mailer in the boilerplate: in development the reset link is
		// printed to the server log.
		h.log.Info("password reset link generated", "email", email, "reset_url", h.resetURL(r, token))
	}
	webctx.SetFlash(w, "success", "If that email belongs to an account, a reset link is on its way. It expires in one hour.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) renderForgot(w http.ResponseWriter, r *http.Request, d templates.ForgotData, code int) {
	d.Title = "Forgot password | Bara"
	d.Theme = webctx.Theme(r)
	d.CSRF = webctx.CSRFToken(r.Context())
	if d.Errors == nil {
		d.Errors = map[string]string{}
	}
	if code != http.StatusOK {
		w.WriteHeader(code)
	}
	components.AuthLayout(h.authData(r, "forgot", templates.ForgotCard(d))).Render(r.Context(), w)
}

func (h *Handler) resetForm(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	d := templates.ResetData{
		Token:  token,
		Errors: map[string]string{},
		Valid:  h.svc.ValidResetToken(token),
	}
	h.renderReset(w, r, d, http.StatusOK)
}

func (h *Handler) resetSubmit(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	if !isHex64(token) {
		h.renderReset(w, r, templates.ResetData{Token: token, Valid: false}, http.StatusBadRequest)
		return
	}

	errs := map[string]string{}
	switch {
	case password == "":
		errs["password"] = "Password is required."
	case !ValidPassword(password):
		errs["password"] = "Use at least 8 characters."
	}
	if confirm != password {
		errs["confirm_password"] = "Passwords don't match."
	}
	if len(errs) > 0 {
		h.renderReset(w, r, templates.ResetData{Token: token, Valid: true, Errors: errs}, http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.ResetPassword(token, password); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			h.renderReset(w, r, templates.ResetData{Token: token, Valid: false}, http.StatusBadRequest)
			return
		}
		h.log.Error("password reset failed", "error", err)
		h.renderReset(w, r, templates.ResetData{Token: token, Valid: true, Errors: map[string]string{
			"form": "Something went wrong on our side. Try again.",
		}}, http.StatusInternalServerError)
		return
	}

	webctx.SetFlash(w, "success", "Password updated. Sign in with your new password.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) renderReset(w http.ResponseWriter, r *http.Request, d templates.ResetData, code int) {
	d.Title = "Reset password | Bara"
	d.Theme = webctx.Theme(r)
	d.CSRF = webctx.CSRFToken(r.Context())
	if d.Errors == nil {
		d.Errors = map[string]string{}
	}
	if code != http.StatusOK {
		w.WriteHeader(code)
	}
	components.AuthLayout(h.authData(r, "forgot", templates.ResetCard(d))).Render(r.Context(), w)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if sess := webctx.Session(r.Context()); sess != nil {
		h.sessions.Destroy(sess.ID)
	}
	webctx.ClearSessionCookie(w, h.cfg.SecureCookies)
	webctx.SetFlash(w, "info", "You have signed out.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request, user *domain.User, msg string) {
	sess, err := h.sessions.Create(user.ID, r.UserAgent(), middleware.ClientIP(r))
	if err != nil {
		h.log.Error("create session failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	webctx.SetSessionCookie(w, h.cfg.SecureCookies, sess.ID, h.sessions.TTL())
	webctx.SetFlash(w, "success", msg)
	target := "/dashboard"
	if user.IsAdmin() {
		target = "/admin"
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (h *Handler) resetURL(r *http.Request, token string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/reset-password?token=%s", scheme, r.Host, token)
}
