package uploads

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// webpFixture returns a minimal but magic-valid WebP payload of size n.
func webpFixture(n int) []byte {
	b := make([]byte, n)
	copy(b, "RIFF\x00\x00\x00\x00WEBP")
	return b
}

func TestSaveAcceptsWebP(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	url, err := s.Save("avatar", webpFixture(100))
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !strings.HasPrefix(url, "/uploads/avatar/") || !strings.HasSuffix(url, ".webp") {
		t.Fatalf("unexpected URL path %q", url)
	}
	name := strings.TrimPrefix(url, "/uploads/avatar/")
	f, err := s.Open("avatar", name)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 100 || string(got[:12]) != string(webpFixture(12)[:12]) {
		t.Fatal("stored content differs from uploaded content")
	}
}

func TestSaveRejectsNonWebP(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	for name, data := range map[string][]byte{
		"png":   {0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0, 0},
		"jpeg":  {0xff, 0xd8, 0xff, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		"text":  []byte("hello world!"),
		"short": {'R', 'I'},
	} {
		if _, err := s.Save("avatar", data); !errors.Is(err, ErrNotWebP) {
			t.Fatalf("%s: expected ErrNotWebP, got %v", name, err)
		}
	}
}

func TestSaveRejectsUnknownTypeAndOversize(t *testing.T) {
	s := NewStore(t.TempDir(), 64)
	if _, err := s.Save("banner", webpFixture(32)); !errors.Is(err, ErrUnknownType) {
		t.Fatalf("expected ErrUnknownType, got %v", err)
	}
	if _, err := s.Save("avatar", webpFixture(65)); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
	if _, err := s.Save("avatar", webpFixture(64)); err != nil {
		t.Fatalf("exact-limit file should be accepted: %v", err)
	}
}

func TestOpenRejectsTraversalAndBadNames(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	for _, name := range []string{
		"../secret.webp",
		"..%2fsecret.webp",
		"avatar.webp",
		"not-a-uuid.webp",
		"12345678-1234-4234-8234-123456789012.png",
		"",
		".webp",
	} {
		if _, err := s.Open("avatar", name); !errors.Is(err, ErrBadName) {
			t.Fatalf("%q: expected ErrBadName, got %v", name, err)
		}
	}
	if _, err := s.Open("avatar", "12345678-1234-4234-8234-123456789012.webp"); !os.IsNotExist(err) {
		t.Fatalf("missing file should report not-exist, got %v", err)
	}
	// A file planted outside the asset dir must stay unreachable.
	outside := filepath.Join(s.dir, "planted.webp")
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, webpFixture(20), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Open("avatar", "planted.webp"); !errors.Is(err, ErrBadName) {
		t.Fatalf("planted file reachable: %v", err)
	}
}

func TestDeleteRemovesFile(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	url, err := s.Save("avatar", webpFixture(40))
	if err != nil {
		t.Fatal(err)
	}
	name := strings.TrimPrefix(url, "/uploads/avatar/")
	if err := s.Delete("avatar", name); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Open("avatar", name); !os.IsNotExist(err) {
		t.Fatalf("file should be gone, got %v", err)
	}
	if err := s.Delete("avatar", "../x.webp"); !errors.Is(err, ErrBadName) {
		t.Fatalf("expected ErrBadName, got %v", err)
	}
}

func TestSavedNamesAreUniqueUUIDs(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	seen := map[string]bool{}
	for i := 0; i < 5; i++ {
		url, err := s.Save("avatar", webpFixture(20))
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimPrefix(url, "/uploads/avatar/")
		if !uuidFilePattern.MatchString(name) {
			t.Fatalf("%q is not a uuid.webp name", name)
		}
		if seen[name] {
			t.Fatal("duplicate file name generated")
		}
		seen[name] = true
	}
}
