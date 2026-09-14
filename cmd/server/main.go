// Command server runs the parameters-network REST API.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	corelog "github.com/jasonmiller-cc/parameters-core/pkg/log"
	"github.com/jasonmiller-cc/parameters-core/pkg/middleware"
	coreserver "github.com/jasonmiller-cc/parameters-core/pkg/server"
	"github.com/jasonmiller-cc/parameters-network/internal/api"
	"github.com/jasonmiller-cc/parameters-network/internal/config"
	"github.com/jasonmiller-cc/parameters-network/internal/service"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		panic("load config: " + err.Error())
	}

	log := corelog.New(corelog.LevelInfo, corelog.FormatJSON)
	log.Info("starting parameters-network", "addr", cfg.Server.Addr())

	svc := service.NewNetworkService()
	h := api.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	handler := middleware.RequestID(middleware.Logger(log)(mux))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := coreserver.New(cfg.Server, handler, log)
	if err := srv.Run(ctx); err != nil {
		log.Error("server error", "err", err)
	}
}
