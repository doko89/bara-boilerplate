package components

import (
	"testing"

	"boilerplate/internal/domain"
)

// TestHomeHref locks the breadcrumb root per role: admins land on the admin
// overview, everyone else on the user dashboard.
func TestHomeHref(t *testing.T) {
	if got := HomeHref(domain.RoleAdmin); got != "/admin" {
		t.Errorf("HomeHref(admin) = %q, want %q", got, "/admin")
	}
	if got := HomeHref(domain.RoleUser); got != "/dashboard" {
		t.Errorf("HomeHref(user) = %q, want %q", got, "/dashboard")
	}
}
