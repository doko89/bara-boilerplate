// Package uploads stores user-supplied image files on local disk and serves
// them back at /uploads/<assetType>/<uuid>.webp. Only the asset types in
// ValidTypes are accepted and every file is validated as WebP before it is
// stored, so a stored file can always be served as image/webp.
package uploads

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var (
	ErrUnknownType = errors.New("unknown asset type")
	ErrTooLarge    = errors.New("file exceeds the size limit")
	ErrNotWebP     = errors.New("file is not a WebP image")
	ErrBadName     = errors.New("invalid file name")
)

// ValidTypes whitelists the asset types that may appear in /uploads URLs.
// The client converts every avatar to WebP (canvas.toBlob, quality 0.8)
// before upload, so the store only has to accept WebP.
var ValidTypes = map[string]bool{"avatar": true}

var uuidFilePattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\.webp$`)

// RoutePrefix is the public URL prefix this feature serves.
const RoutePrefix = "/uploads/"

type Store struct {
	dir      string
	maxBytes int64
}

func NewStore(dir string, maxBytes int64) *Store {
	return &Store{dir: dir, maxBytes: maxBytes}
}

func (s *Store) MaxBytes() int64 { return s.maxBytes }

// URLPath returns the public path for a stored file.
func URLPath(assetType, name string) string {
	return RoutePrefix + assetType + "/" + name
}

// IsWebP reports whether data looks like a WebP image (RIFF....WEBP).
func IsWebP(data []byte) bool {
	return len(data) >= 12 &&
		data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P'
}

// Save validates data and stores it as a fresh <uuid>.webp file under the
// asset type directory. It returns the public URL path of the stored file.
func (s *Store) Save(assetType string, data []byte) (string, error) {
	if !ValidTypes[assetType] {
		return "", ErrUnknownType
	}
	if int64(len(data)) > s.maxBytes {
		return "", ErrTooLarge
	}
	if !IsWebP(data) {
		return "", ErrNotWebP
	}
	name, err := newUUID()
	if err != nil {
		return "", err
	}
	name += ".webp"
	dir := filepath.Join(s.dir, assetType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("uploads: create directory: %w", err)
	}
	// The name is generated above, never taken from the request, so the
	// joined path cannot escape the asset directory.
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", fmt.Errorf("uploads: write file: %w", err)
	}
	return URLPath(assetType, name), nil
}

// Open returns the stored file for serving after validating the asset type
// and the uuid.webp file name. The strict name pattern keeps the joined
// path inside the asset directory (no traversal possible).
func (s *Store) Open(assetType, name string) (*os.File, error) {
	if !ValidTypes[assetType] {
		return nil, ErrUnknownType
	}
	if !uuidFilePattern.MatchString(name) {
		return nil, ErrBadName
	}
	f, err := os.Open(filepath.Join(s.dir, assetType, name))
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Delete removes a previously stored file. It validates its inputs the
// same way Open does and reports whether a file was actually removed.
func (s *Store) Delete(assetType, name string) error {
	if !ValidTypes[assetType] {
		return ErrUnknownType
	}
	if !uuidFilePattern.MatchString(name) {
		return ErrBadName
	}
	if err := os.Remove(filepath.Join(s.dir, assetType, name)); err != nil {
		return err
	}
	return nil
}

// newUUID returns a random UUID v4 string using only the standard library.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("uploads: random uuid: %w", err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
