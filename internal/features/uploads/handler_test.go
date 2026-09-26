package uploads

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testMux(t *testing.T, s *Store) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	RegisterRoutes(mux, s)
	return mux
}

func TestServeStoredFile(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	data := webpFixture(64)
	url, err := s.Save("avatar", data)
	if err != nil {
		t.Fatal(err)
	}
	mux := testMux(t, s)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("content-type = %q, want image/webp", ct)
	}
	got, _ := io.ReadAll(res.Body)
	if string(got) != string(data) {
		t.Fatal("served bytes differ from stored bytes")
	}
}

func TestServeNotFoundCases(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	mux := testMux(t, s)
	for name, path := range map[string]string{
		"missing file": "/uploads/avatar/12345678-1234-4234-8234-123456789012.webp",
		"bad type":     "/uploads/banner/12345678-1234-4234-8234-123456789012.webp",
		"bad name":     "/uploads/avatar/not-a-uuid.webp",
		"traversal":    "/uploads/avatar/..%2f..%2fsecret.webp",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s: status = %d, want 404", name, rec.Code)
		}
	}
}

func TestServePathValueWiring(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	url, err := s.Save("avatar", webpFixture(32))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "/uploads/avatar/") {
		t.Fatalf("unexpected URL %q", url)
	}
}
