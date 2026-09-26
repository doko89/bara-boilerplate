package uploads

import (
	"errors"
	"net/http"
	"os"
)

// RegisterRoutes serves stored files at /uploads/<assetType>/<uuid>.webp.
// Reads are public and cacheable for a year: file names are content
// addressed (a fresh uuid per upload), so a URL never changes meaning.
func RegisterRoutes(mux *http.ServeMux, store *Store) {
	h := &Handler{store: store}
	mux.HandleFunc("GET /uploads/{type}/{file}", h.serve)
}

type Handler struct {
	store *Store
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request) {
	f, err := h.store.Open(r.PathValue("type"), r.PathValue("file"))
	if err != nil {
		if errors.Is(err, ErrUnknownType) || errors.Is(err, ErrBadName) {
			http.NotFound(w, r)
			return
		}
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "cannot read file", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		http.Error(w, "cannot read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}
