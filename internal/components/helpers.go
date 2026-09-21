package components

import (
	"fmt"
	"strings"
	"time"

	"boilerplate/internal/domain"

	"github.com/a-h/templ"
)

// AppPageData wires everything the authenticated app shell needs.
type AppPageData struct {
	Title   string
	Active  string // id of the active sidebar item
	Theme   string // "light", "dark" or "" (follow system)
	User    *domain.User
	CSRF    string
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

func NavItems(role domain.Role) []NavItem {
	overview := NavItem{ID: "overview", Label: "Overview", Href: "/dashboard", Icon: "home"}
	if role == domain.RoleAdmin {
		overview.Href = "/admin"
	}
	items := []NavItem{overview}
	if role == domain.RoleAdmin {
		items = append(items, NavItem{ID: "users", Label: "Users", Href: "/admin/users", Icon: "users"})
	}
	return append(items, NavItem{ID: "settings", Label: "Settings", Href: "/settings", Icon: "sliders"})
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
