package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWithNotFoundKeepsPathValues guards the wildcard routing contract:
// requests that match a parameterized pattern must reach the handler with
// their path values populated, while unknown paths get the styled 404.
func TestWithNotFoundKeepsPathValues(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /files/{id}", func(w http.ResponseWriter, r *http.Request) {
		if got := r.PathValue("id"); got != "abc123" {
			t.Errorf("PathValue(id) = %q, want %q", got, "abc123")
		}
		w.Write([]byte("id=" + r.PathValue("id")))
	})

	h := withNotFound(mux)

	matched := httptest.NewRequest(http.MethodGet, "/files/abc123", nil)
	matchedRec := httptest.NewRecorder()
	h.ServeHTTP(matchedRec, matched)
	if matchedRec.Code != http.StatusOK {
		t.Fatalf("matched status = %d, want 200", matchedRec.Code)
	}
	if body := matchedRec.Body.String(); body != "id=abc123" {
		t.Fatalf("matched body = %q, want %q", body, "id=abc123")
	}

	missing := httptest.NewRequest(http.MethodGet, "/no/such/page", nil)
	missingRec := httptest.NewRecorder()
	h.ServeHTTP(missingRec, missing)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("unmatched status = %d, want 404", missingRec.Code)
	}
}
