package main

import (
	"log"
	"net/http"

	"github.com/qingketsing/novel2script/backend/internal/app"
	"github.com/qingketsing/novel2script/backend/internal/config"
	httpapi "github.com/qingketsing/novel2script/backend/internal/http"
)

// main 组装配置、路由和占位转换器，启动后端 HTTP 服务。
func main() {
	cfg := config.Load()
	router := newHandler()

	log.Printf("starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatal(err)
	}
}

// newHandler 组装生产路由，并接入 mock 领域转换器作为当前 MVP 的转换管线。
func newHandler() http.Handler {
	return httpapi.NewRouter(app.NewMockDomainConverter())
}
