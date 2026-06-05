package main

import (
	"log"
	"net/http"

	"github.com/qingketsing/novel2script/backend/internal/app"
	"github.com/qingketsing/novel2script/backend/internal/config"
	httpapi "github.com/qingketsing/novel2script/backend/internal/http"
)

func main() {
	cfg := config.Load()
	router := httpapi.NewRouter(app.NewPlaceholderConverter())

	log.Printf("starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatal(err)
	}
}
