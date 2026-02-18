package handler

import (

	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"digio-games/web"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	romsFS fs.FS
}

func New(romsFS fs.FS) *Handler {
	return &Handler{romsFS: romsFS}
}

type romEntry struct {
	Name     string `json:"name"`
	FileName string `json:"file_name"`
}

func (h *Handler) IndexPage(w http.ResponseWriter, r *http.Request) {
	web.Templates().ExecuteTemplate(w, "index.html", map[string]any{
		"ROMs": h.scanROMs(),
	})
}

func (h *Handler) PlayPage(w http.ResponseWriter, r *http.Request) {
	rom := chi.URLParam(r, "rom")
	web.Templates().ExecuteTemplate(w, "play.html", map[string]any{
		"ROM":  rom,
		"Name": prettifyROM(rom),
	})
}

func (h *Handler) ListROMs(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, h.scanROMs())
}

func (h *Handler) ServeROM(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	safe := filepath.Base(name)

	f, err := h.romsFS.Open(safe)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	io.Copy(w, f)
}



func (h *Handler) scanROMs() []romEntry {
	entries, err := fs.ReadDir(h.romsFS, ".")
	if err != nil {
		return nil
	}
	var roms []romEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".sfc" || ext == ".smc" || ext == ".zip" {
			roms = append(roms, romEntry{
				Name:     prettifyROM(e.Name()),
				FileName: e.Name(),
			})
		}
	}
	return roms
}

func prettifyROM(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.NewReplacer("_", " ", "-", " ").Replace(name)
	return name
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errBody(err error) map[string]string {
	return map[string]string{"error": err.Error()}
}
