package profile

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"boilerplate/internal/domain"
	"boilerplate/internal/features/profile/templates"
)

// TestProfileRequiresAuth locks the renamed route: /profile is served by this
// feature and still guarded by the auth check.
func TestProfileRequiresAuth(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusSeeOther)
	}
	if loc := res.Header.Get("Location"); loc != "/login" {
		t.Fatalf("location = %q, want %q", loc, "/login")
	}
}

// TestSettingsRedirectsToProfile keeps the old /settings URL working after
// the rename.
func TestSettingsRedirectsToProfile(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusMovedPermanently)
	}
	if loc := res.Header.Get("Location"); loc != "/profile" {
		t.Fatalf("location = %q, want %q", loc, "/profile")
	}
}

// TestProfileContentFormActions guards the template side of the rename: every
// form on the page must post to a registered /profile route, and the avatar
// dropzone must keep its upload wiring.
func TestProfileContentFormActions(t *testing.T) {
	var sb strings.Builder
	d := templates.ProfileData{
		Errors:   map[string]string{},
		Sessions: []domain.Session{{ID: "a"}, {ID: "b"}},
	}
	if err := templates.ProfileContent(d).Render(context.Background(), &sb); err != nil {
		t.Fatal(err)
	}
	html := sb.String()
	for _, want := range []string{
		`action="/profile"`,
		`action="/profile/password"`,
		`action="/profile/sessions/revoke-others"`,
		`data-avatar-dropzone`,
		`class="avatar avatar-xl avatar-upload"`,
		`name="avatar"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered profile page missing %q", want)
		}
	}
	if strings.Contains(html, "Cropped square") {
		t.Error("rendered profile page still contains the removed avatar help text")
	}
	// Sections render in visual 2-column order: account (with the avatar),
	// password, then full-width sessions. DOM order matches so keyboard tab
	// order follows the layout.
	if strings.Contains(html, `id="avatar-title"`) {
		t.Error("rendered profile page still contains the removed photo section")
	}
	last := -1
	for _, id := range []string{"profile-title", "password-title", "sessions-title"} {
		i := strings.Index(html, `id="`+id+`"`)
		if i < 0 {
			t.Fatalf("rendered profile page missing %q", id)
		}
		if i < last {
			t.Fatalf("section %q renders out of order", id)
		}
		last = i
	}
}
