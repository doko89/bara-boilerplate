package components

import (
	"context"
	"strings"
	"testing"

	"boilerplate/internal/domain"
)

// TestBaseVersionsStaticAssets locks cache busting: every local CSS/JS URL
// carries ?v=StaticVersion so browsers fetch fresh bytes after a deploy
// instead of rendering the cached copy.
func TestBaseVersionsStaticAssets(t *testing.T) {
	var sb strings.Builder
	err := Base("t", "", "", []string{"/static/css/app.css"}, []string{"/static/js/app.js"}, Icon("check", 16)).Render(context.Background(), &sb)
	if err != nil {
		t.Fatal(err)
	}
	html := sb.String()
	for _, want := range []string{
		`/static/css/base.css?v=` + StaticVersion,
		`/static/css/app.css?v=` + StaticVersion,
		`/static/js/theme.js?v=` + StaticVersion,
		`/static/js/app.js?v=` + StaticVersion,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered base layout missing %q", want)
		}
	}
}

// TestBreadcrumbRendersTrail locks the breadcrumb contract: earlier items are
// links, the last item is the current page and never a link.
func TestBreadcrumbRendersTrail(t *testing.T) {
	var sb strings.Builder
	err := Breadcrumb([]Crumb{{Label: "Home", Href: "/admin", Icon: "home"}, {Label: "Admin overview", Href: "/admin"}, {Label: "Users"}}).Render(context.Background(), &sb)
	if err != nil {
		t.Fatal(err)
	}
	html := sb.String()
	for _, want := range []string{
		`aria-label="Breadcrumb"`,
		`<svg class="icon" width="14"`,
		`<span class="sr-only">Home</span>`,
		`>Admin overview</a>`,
		`<span aria-current="page">Users</span>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered breadcrumb missing %q", want)
		}
	}
	if strings.Contains(html, `<a href="">`) {
		t.Error("rendered breadcrumb links the current page")
	}
}

// TestAppShellRendersTopbarBreadcrumb locks the topbar contract: when crumbs
// are set the breadcrumb trail replaces the plain title (kept screen-reader
// only), otherwise the visible h1 title renders.
func TestAppShellRendersTopbarBreadcrumb(t *testing.T) {
	render := func(d AppPageData) string {
		var sb strings.Builder
		if err := AppShell(d).Render(context.Background(), &sb); err != nil {
			t.Fatal(err)
		}
		return sb.String()
	}
	user := &domain.User{ID: 1, Name: "Root Admin", Email: "admin@bara.dev", Role: domain.RoleAdmin}

	withCrumbs := render(AppPageData{
		Title:   "Users",
		Active:  "users",
		User:    user,
		CSRF:    "csrf",
		Crumbs:  []Crumb{{Label: "Home", Href: "/admin", Icon: "home"}, {Label: "Users"}},
		Content: Icon("check", 16),
	})
	for _, want := range []string{
		`aria-label="Breadcrumb"`,
		`<h1 class="sr-only">Users</h1>`,
		`<span aria-current="page">Users</span>`,
	} {
		if !strings.Contains(withCrumbs, want) {
			t.Errorf("topbar with crumbs missing %q", want)
		}
	}
	if strings.Contains(withCrumbs, "topbar-title") {
		t.Error("topbar with crumbs still renders the plain title")
	}

	plain := render(AppPageData{
		Title:   "Users",
		Active:  "users",
		User:    user,
		CSRF:    "csrf",
		Content: Icon("check", 16),
	})
	if !strings.Contains(plain, `<h1 class="topbar-title">Users</h1>`) {
		t.Error("topbar without crumbs missing the plain title")
	}
	if strings.Contains(plain, `aria-label="Breadcrumb"`) {
		t.Error("topbar without crumbs renders a breadcrumb")
	}
}
