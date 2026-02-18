package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"digio-games/internal/handler"
	"digio-games/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed roms
var romsFS embed.FS

func main() {
	romsSubFS, err := fs.Sub(romsFS, "roms")
	if err != nil {
		log.Fatalf("roms embed: %v", err)
	}

	h := handler.New(romsSubFS)

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
			r.Head("/roms/{name}", h.ServeROM)
			r.Get("/roms/{name}", h.ServeROM)

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
