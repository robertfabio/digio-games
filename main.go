package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"digio-games/internal/db"
	"digio-games/internal/handler"
	"digio-games/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	store, err := db.NewStore(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	romsDir := envOr("ROMS_DIR", "roms")
	os.MkdirAll(romsDir, 0o755)

	h := handler.New(store, romsDir)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/digio.com.br/", http.StatusFound)
	})

	r.Route("/digio.com.br", func(r chi.Router) {
		r.Get("/", h.IndexPage)
		r.Get("/play/{rom}", h.PlayPage)

		r.Route("/api", func(r chi.Router) {
			r.Get("/roms", h.ListROMs)
			r.Get("/roms/{name}", h.ServeROM)
			r.Get("/saves/{rom}", h.ListSaves)
			r.Get("/saves/{rom}/data", h.GetSaveData)
			r.Post("/saves/{rom}", h.SaveGame)
			r.Delete("/saves/{rom}/{id}", h.DeleteSave)
		})

		r.Handle("/static/*", http.StripPrefix("/digio.com.br/static/", http.FileServerFS(web.StaticFS())))
	})

	port := envOr("PORT", "8080")
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
