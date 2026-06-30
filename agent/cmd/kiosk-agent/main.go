package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
	"github.com/aschorle/kiosk-client/agent/internal/status"
	"github.com/aschorle/kiosk-client/agent/internal/web"
)

const (
	version          = "0.5.0"
	configPath       = "config/client.conf"
	httpAddr         = ":8080"
	watchdogInterval = 30 * time.Second
)

func main() {
	log.Printf("starting kiosk-agent %s", version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalChannel)

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

	watchdogDone := browser.StartWatchdog(ctx, watchdogInterval, log.Printf)
	log.Printf("browser watchdog started with interval %s", watchdogInterval)
	log.Printf("http server listening on %s", httpAddr)

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.ListenAndServe()
	}()

	select {
	case err := <-serverDone:
		log.Printf("http server stopped: %v", err)
	case signal := <-signalChannel:
		log.Printf("received signal: %s", signal)
	}

	cancel()
	<-watchdogDone
}
