package main

import (
	"log"

	"github.com/aschorle/kiosk-client/agent/internal/config"
	"github.com/aschorle/kiosk-client/agent/internal/status"
	"github.com/aschorle/kiosk-client/agent/internal/web"
)

const (
	version    = "0.4.0"
	configPath = "config/client.conf"
	httpAddr   = ":8080"
)

func main() {
	log.Printf("starting kiosk-agent %s", version)

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	provider := status.NewProvider(cfg, version)
	server := web.NewServer(httpAddr, provider)

	log.Printf("configuration loaded: url=%s device_id=%s browser=%s", cfg.URL, cfg.DeviceID, cfg.Browser)
	for _, route := range server.Routes() {
		log.Printf("registered route: %s %s", route.Method, route.Path)
	}
	log.Printf("http server listening on %s", httpAddr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("http server stopped: %v", err)
	}
}
