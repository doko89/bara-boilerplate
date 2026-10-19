package components

import (
	"fmt"
	"strings"
	"time"

	"boilerplate/internal/domain"

	"github.com/a-h/templ"
)

// StaticVersion is appended to every local CSS/JS URL as ?v=. Static files
// are served with a long Cache-Control, so bump this whenever anything under
// web/static changes — otherwise browsers keep rendering the cached bytes.
const StaticVersion = "10"

// AppPageData wires everything the authenticated app shell needs.
type AppPageData struct {
	Title   string
	Active  string // id of the active sidebar item
	Crumbs  []Crumb
	Theme   string // "light", "dark" or "" (follow system)
	User    *domain.User
	CSRF    string
	Scripts []string // extra page scripts besides app.js, e.g. avatar.js
	Content templ.Component
}

// AuthPageData wires the asymmetric split auth layout.
type AuthPageData struct {
	Title      string
	Theme      string
	User       *domain.User
	CSRF       string
	PanelTitle string
	PanelText  string
	Content    templ.Component
}

type NavItem struct {
	ID    string
	Label string
	Href  string
	Icon  string
}

// Crumb is one breadcrumb item. The last crumb is rendered as the current
// page; Href on it is ignored. When Icon is set the label is visually
// replaced by the icon (kept for screen readers).
type Crumb struct {
	Label string
	Href  string
	Icon  string
}

type ButtonVariant string

const (
	ButtonPrimary   ButtonVariant = "primary"
	ButtonSecondary ButtonVariant = "secondary"
	ButtonInverse   ButtonVariant = "inverse"
)

type ButtonSize string

const (
	ButtonMD ButtonSize = "md"
	ButtonSM ButtonSize = "sm"
	ButtonLG ButtonSize = "lg"
)

// ButtonProps describes a themed button. Href renders an <a> instead of a
// <button> (Type and Disabled then don't apply). Zero values render a medium
// primary submit button.
type ButtonProps struct {
	Label    string
	Href     string
	Type     string
	Variant  ButtonVariant
	Size     ButtonSize
	Block    bool
	Disabled bool
	Icon     string
}

func buttonType(p ButtonProps) string {
	if p.Type != "" {
		return p.Type
	}
	return "submit"
}

func buttonClass(p ButtonProps) string {
	v := p.Variant
	if v == "" {
		v = ButtonPrimary
	}
	cls := "btn btn-" + string(v)
	switch p.Size {
	case ButtonSM:
		cls += " btn-sm"
	case ButtonLG:
		cls += " btn-lg"
	}
	if p.Block {
		cls += " btn-block"
	}
	return cls
}

type DropdownAlign string

const (
	DropdownDown  DropdownAlign = ""
	DropdownUp    DropdownAlign = "up"
	DropdownRight DropdownAlign = "right"
)

// DropdownItem is one link in a Dropdown menu. Selected marks the current
// choice (e.g. the active filter).
type DropdownItem struct {
	Label    string
	Href     string
	Icon     string
	Selected bool
}

// DropdownProps describes a toggle button with a link menu, for selects such
// as filters. Behavior comes from the shared data-dropdown script.
type DropdownProps struct {
	Label string
	Align DropdownAlign
	Items []DropdownItem
}

func NavItems(role domain.Role) []NavItem {
	overview := NavItem{ID: "overview", Label: "Overview", Href: "/dashboard", Icon: "home"}
	if role == domain.RoleAdmin {
		overview.Href = "/admin"
	}
	items := []NavItem{overview}
	if role == domain.RoleAdmin {
		items = append(items, NavItem{ID: "users", Label: "Users", Href: "/admin/users", Icon: "users"})
	}
	// Profile lives only in the account dropdown (sidebar footer), not here.
	return items
}

// HomeHref returns the landing page for a role: the admin overview for
// admins, the user dashboard otherwise. Breadcrumb trails root here.
func HomeHref(role domain.Role) string {
	if role == domain.RoleAdmin {
		return "/admin"
	}
	return "/dashboard"
}

// AuthPanelCopy returns the marketing copy shown on the brand panel of each
// auth page.
func AuthPanelCopy(page string) (title, text string) {
	switch page {
	case "register":
		return "Start on a stack that already works.",
		"Create an account and look around. Both the user and the admin side are included."
	case "forgot":
		return "Locked out? It happens.",
		"Tell us the email on the account and we will send a single-use reset link that expires in one hour."
	default:
		return "Pick up right where you left off.",
		"Sessions, roles and security defaults are already wired. Sign in and see both sides of the app."
	}
}

func ErrorPageTitle(code int) string {
	return fmt.Sprintf("%d | Bara", code)
}

func TimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2, 2006")
	}
}

func FormatDate(t time.Time) string { return t.Format("Jan 2, 2006") }

// FirstName returns the first word of a display name.
func FirstName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "there"
	}
	if i := strings.IndexAny(name, " \t"); i > 0 {
		return name[:i]
	}
	return name
}

func Greeting(now time.Time) string {
	switch h := now.Hour(); {
	case h < 12:
		return "Good morning"
	case h < 18:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

func DeviceLabel(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "iphone"):
		return "iPhone"
	case strings.Contains(ua, "ipad"):
		return "iPad"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "mac os"):
		return "Mac"
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "linux"):
		return "Linux"
	default:
		return "Unknown device"
	}
}

func BrowserLabel(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "edg/"):
		return "Edge"
	case strings.Contains(ua, "chrome"), strings.Contains(ua, "crios"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari"):
		return "Safari"
	default:
		return "Browser"
	}
}

func SessionLabel(userAgent string) string {
	return BrowserLabel(userAgent) + " on " + DeviceLabel(userAgent)
}
