package components

import (
	"context"
	"strings"
	"testing"
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
