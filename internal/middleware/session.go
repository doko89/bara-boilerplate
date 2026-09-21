package middleware

import (
	"net/http"
	"time"

	"boilerplate/internal/session"
	"boilerplate/internal/webctx"
)

// LoadSession resolves the HttpOnly session cookie into the current user and
// stores both in the request context.
func LoadSession(manager *session.Manager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c, err := r.Cookie(webctx.SessionCookieName); err == nil && c.Value != "" {
				if s, err := manager.Get(c.Value); err == nil {
					ctx := webctx.WithUser(r.Context(), &s.User)
					ctx = webctx.WithSession(ctx, s)
					r = r.WithContext(ctx)
					if time.Since(s.LastSeenAt) > 5*time.Minute {
						manager.Touch(s)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func IsAuthenticated(r *http.Request) bool {
	return webctx.User(r.Context()) != nil
}

// RequireAuth is a per-route guard: call it at the top of a handler and bail
// out when it returns false.
func RequireAuth(w http.ResponseWriter, r *http.Request) bool {
	if webctx.User(r.Context()) == nil {
		webctx.SetFlash(w, "error", "Please sign in to continue.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	return true
}

func RequireAdmin(w http.ResponseWriter, r *http.Request) bool {
	u := webctx.User(r.Context())
	if u == nil {
		webctx.SetFlash(w, "error", "Please sign in to continue.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	if !u.IsAdmin() {
		webctx.SetFlash(w, "error", "That area is only for admins.")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return false
	}
	return true
}

// RedirectIfAuthenticated keeps signed-in users away from auth pages.
func RedirectIfAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	if webctx.User(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return true
	}
	return false
}
