package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"

	"boilerplate/internal/components"
	"boilerplate/internal/webctx"
)

const CSRFFieldName = "_csrf"

const csrfTokenTTLSeconds = 12 * 3600

// NewCSRF implements double-submit cookie CSRF protection with an origin
// check. The token lives in an HttpOnly cookie; forms embed the same token as
// a hidden field (components.CSRFField). All unsafe methods are verified.
func NewCSRF(secureCookies bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if c, err := r.Cookie(webctx.CSRFCookieName); err == nil && isToken(c.Value) {
				token = c.Value
			}

			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				// Issue a token on first visit so rendered forms carry one.
				if token == "" && !strings.HasPrefix(r.URL.Path, "/static/") {
					token = newToken()
					http.SetCookie(w, &http.Cookie{
						Name:     webctx.CSRFCookieName,
						Value:    token,
						Path:     "/",
						MaxAge:   csrfTokenTTLSeconds,
						HttpOnly: true,
						Secure:   secureCookies,
						SameSite: http.SameSiteLaxMode,
					})
				}
			default:
				supplied := r.Header.Get("X-CSRF-Token")
				if supplied == "" {
					if err := r.ParseForm(); err == nil {
						supplied = r.PostForm.Get(CSRFFieldName)
					}
				}
				if token == "" || supplied == "" ||
					subtle.ConstantTimeCompare([]byte(token), []byte(supplied)) != 1 {
					renderBlocked(w, r)
					return
				}
				if !sameOrigin(r) {
					renderBlocked(w, r)
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(webctx.WithCSRFToken(r.Context(), token)))
		})
	}
}

func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("csrf: cannot generate token: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func isToken(v string) bool {
	if len(v) != 64 {
		return false
	}
	_, err := hex.DecodeString(v)
	return err == nil
}

// sameOrigin rejects cross-site form posts even if the token leaked.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // non-browser clients
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(u.Host), []byte(r.Host)) == 1
}

func renderBlocked(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	components.ErrorPage(
		http.StatusForbidden,
		"Request blocked",
		"The form token was missing, expired or did not match. Go back, reload the page and try again.",
		webctx.Theme(r),
		webctx.User(r.Context()),
	).Render(r.Context(), w)
}
