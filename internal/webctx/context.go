// Package webctx carries per-request data (current user, session, CSRF token)
// through context and centralises cookie helpers.
package webctx

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"boilerplate/internal/domain"
)

type contextKey int

const (
	userKey contextKey = iota + 1
	sessionKey
	csrfKey
)

const (
	SessionCookieName = "bara_session"
	CSRFCookieName    = "bara_csrf"
	ThemeCookieName   = "bara_theme"
	FlashCookieName   = "bara_flash"
)

func WithUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func User(ctx context.Context) *domain.User {
	u, _ := ctx.Value(userKey).(*domain.User)
	return u
}

func WithSession(ctx context.Context, s *domain.Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func Session(ctx context.Context) *domain.Session {
	s, _ := ctx.Value(sessionKey).(*domain.Session)
	return s
}

func WithCSRFToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfKey, token)
}

func CSRFToken(ctx context.Context) string {
	t, _ := ctx.Value(csrfKey).(string)
	return t
}

// Theme reads the persisted theme choice. An empty return value means "follow
// the system preference".
func Theme(r *http.Request) string {
	c, err := r.Cookie(ThemeCookieName)
	if err != nil {
		return ""
	}
	switch c.Value {
	case "light", "dark":
		return c.Value
	}
	return ""
}

type Flash struct {
	Type    string `json:"t"` // "success" | "error" | "info"
	Message string `json:"m"`
}

func SetFlash(w http.ResponseWriter, flashType, message string) {
	data, err := json.Marshal(Flash{Type: flashType, Message: message})
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     FlashCookieName,
		Value:    base64.URLEncoding.EncodeToString(data),
		Path:     "/",
		MaxAge:   120,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ConsumeFlash reads the flash cookie once and clears it.
func ConsumeFlash(w http.ResponseWriter, r *http.Request) (string, string) {
	c, err := r.Cookie(FlashCookieName)
	if err != nil {
		return "", ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     FlashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	raw, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		return "", ""
	}
	var f Flash
	if err := json.Unmarshal(raw, &f); err != nil {
		return "", ""
	}
	return f.Type, f.Message
}

func SetSessionCookie(w http.ResponseWriter, secure bool, id string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
