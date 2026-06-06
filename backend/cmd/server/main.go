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
	router, err := newHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatal(err)
	}
}

// newHandler 根据配置组装生产路由，并把 AI provider 注入领域转换器。
func newHandler(cfg config.Config) (http.Handler, error) {
	provider, err := app.NewProviderFromConfig(cfg)
	if err != nil {
		if cfg.AIFallbackToMock && cfg.AIMode == "deepseek" {
			return httpapi.NewRouter(app.NewMockDomainConverter()), nil
		}
		return nil, err
	}

	converter := app.NewDomainConverter(provider)
	if cfg.AIFallbackToMock && cfg.AIMode == "deepseek" {
		converter = app.NewFallbackConverter(converter, app.NewMockDomainConverter())
	}
	return httpapi.NewRouter(converter), nil
}
