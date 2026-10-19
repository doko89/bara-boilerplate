package admin

import (
	"context"
	"strings"
	"testing"

	"boilerplate/internal/domain"
	"boilerplate/internal/features/admin/templates"
)

// TestUsersContentRowActions locks the no-dropdown row actions: every other
// user gets direct role/status/delete buttons, while the admin's own row gets
// a "You" marker instead of actions.
func TestUsersContentRowActions(t *testing.T) {
	d := templates.UsersData{
		Users: []domain.User{
			{ID: 1, Name: "Root Admin", Email: "admin@bara.dev", Role: domain.RoleAdmin, IsActive: true},
			{ID: 2, Name: "Jane Doe", Email: "jane@example.com", Role: domain.RoleUser, IsActive: true},
		},
		CurrentUserID: 1,
	}
	var sb strings.Builder
	if err := templates.UsersContent(d).Render(context.Background(), &sb); err != nil {
		t.Fatal(err)
	}
	html := sb.String()
	for _, want := range []string{
		`class="row-actions"`,
		`>Action</th>`,
		`action="/admin/users/2/role"`,
		`action="/admin/users/2/status"`,
		`action="/admin/users/2/delete"`,
		`class="icon-btn danger"`,
		`>You<`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered users page missing %q", want)
		}
	}
	for _, gone := range []string{
		`data-dropdown-toggle`,
		`dropdown-item`,
		`dropdown-note`,
	} {
		if strings.Contains(html, gone) {
			t.Errorf("rendered users page still contains %q", gone)
		}
	}
}
