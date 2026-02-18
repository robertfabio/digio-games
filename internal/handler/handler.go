package handler

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"digio-games/internal/db"
	"digio-games/web"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store  *db.Store
	romsFS fs.FS
}

func New(store *db.Store, romsFS fs.FS) *Handler {
	return &Handler{store: store, romsFS: romsFS}
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

func (h *Handler) ListSaves(w http.ResponseWriter, r *http.Request) {
	rom := chi.URLParam(r, "rom")
	saves, err := h.store.ListSaves(r.Context(), rom)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	if saves == nil {
		saves = []db.Save{}
	}
	respondJSON(w, http.StatusOK, saves)
}

func (h *Handler) GetSaveData(w http.ResponseWriter, r *http.Request) {
	rom := chi.URLParam(r, "rom")
	saveType := r.URL.Query().Get("type")
	if saveType == "" {
		saveType = "sram"
	}
	slot, _ := strconv.Atoi(r.URL.Query().Get("slot"))

	data, err := h.store.GetSaveData(r.Context(), rom, saveType, slot)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"data": base64.StdEncoding.EncodeToString(data),
	})
}

type saveRequest struct {
	SaveType string `json:"save_type"`
	Slot     int    `json:"slot"`
	Data     string `json:"data"`
}

func (h *Handler) SaveGame(w http.ResponseWriter, r *http.Request) {
	rom := chi.URLParam(r, "rom")

	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	if req.SaveType == "" {
		req.SaveType = "sram"
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid base64"})
		return
	}

	if err := h.store.UpsertSave(r.Context(), rom, req.SaveType, req.Slot, data); err != nil {
		respondJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) DeleteSave(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.store.DeleteSave(r.Context(), id); err != nil {
		respondJSON(w, http.StatusInternalServerError, errBody(err))
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
