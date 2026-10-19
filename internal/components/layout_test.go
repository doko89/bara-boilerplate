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

// TestButtonRendersVariants locks the Button contract: zero values render a
// medium primary submit, Href renders a link, and variant/size/block/disabled
// map to the matching classes.
func TestButtonRendersVariants(t *testing.T) {
	render := func(p ButtonProps) string {
		var sb strings.Builder
		if err := Button(p).Render(context.Background(), &sb); err != nil {
			t.Fatal(err)
		}
		return sb.String()
	}

	def := render(ButtonProps{Label: "Save"})
	for _, want := range []string{`<button type="submit"`, `class="btn btn-primary"`, `>Save<`} {
		if !strings.Contains(def, want) {
			t.Errorf("default button missing %q", want)
		}
	}
	if strings.Contains(def, "disabled") {
		t.Error("default button must not be disabled")
	}

	link := render(ButtonProps{Label: "Manage", Href: "/admin/users", Variant: ButtonSecondary, Size: ButtonSM})
	for _, want := range []string{`<a href="/admin/users"`, `class="btn btn-secondary btn-sm"`, `>Manage<`} {
		if !strings.Contains(link, want) {
			t.Errorf("link button missing %q", want)
		}
	}

	full := render(ButtonProps{Label: "Go", Block: true, Disabled: true, Icon: "arrow-right", Type: "button"})
	for _, want := range []string{`type="button"`, `btn-block`, `disabled`, `<svg class="icon"`} {
		if !strings.Contains(full, want) {
			t.Errorf("full button missing %q", want)
		}
	}
}

// TestDropdownRendersMenu locks the Dropdown contract: a toggle button wired
// to the shared data-dropdown script plus a link menu where only the selected
// item is highlighted.
func TestDropdownRendersMenu(t *testing.T) {
	var sb strings.Builder
	p := DropdownProps{
		Label: "Role: all",
		Items: []DropdownItem{
			{Label: "All", Href: "?role=", Selected: true},
			{Label: "Admins", Href: "?role=admin", Icon: "shield"},
		},
	}
	if err := Dropdown(p).Render(context.Background(), &sb); err != nil {
		t.Fatal(err)
	}
	html := sb.String()
	for _, want := range []string{
		`data-dropdown-toggle`,
		`aria-haspopup="menu"`,
		`role="menu"`,
		`>Role: all<`,
		`class="dropdown-item active"`,
		`href="?role=admin"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered dropdown missing %q", want)
		}
	}
	if n := strings.Count(html, "dropdown-item active"); n != 1 {
		t.Errorf("dropdown highlights %d items, want 1", n)
	}
}
