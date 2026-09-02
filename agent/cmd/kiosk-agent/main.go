package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
	"github.com/aschorle/kiosk-client/agent/internal/management"
	"github.com/aschorle/kiosk-client/agent/internal/status"
	"github.com/aschorle/kiosk-client/agent/internal/web"
)

const (
	configPath       = "config/client.conf"
	watchdogInterval = 30 * time.Second
)

func main() {
	configFile := flag.String("config", configPath, "path to client configuration")
	flag.Parse()

	status.SetAgentStartTime(time.Now().UTC())
	log.Printf("starting kiosk-agent %s", status.AgentVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalChannel)

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	// Appliance first boot intentionally tolerates a missing default config and
	// serves the welcome flow. A management-only service always supplies an
	// explicit protected path and must never silently fall back to port 8080.
	if *configFile != configPath {
		if _, err := config.Current(); err != nil {
			log.Fatalf("failed to load explicit configuration: %v", err)
		}
	}

	controller, err := browser.NewController(cfg.BrowserController, cfg.Browser, cfg.BrowserService)
	if err != nil {
		log.Fatalf("failed to create browser controller: %v", err)
	}
	provider := status.NewProvider(cfg, status.AgentVersion, controller)
	server := web.NewServer(cfg.HTTPAddr, provider, controller)

	log.Printf("configuration loaded: url=%s device_id=%s browser=%s client_type=%s controller=%s", cfg.URL, cfg.DeviceID, cfg.Browser, cfg.ClientType, cfg.BrowserController)
	for _, route := range server.Routes() {
		log.Printf("registered route: %s %s", route.Method, route.Path)
	}

	watchdogDone := browser.StartWatchdogWithController(ctx, watchdogInterval, controller, cfg.WatchdogMode, log.Printf)
	watchdogMetricsDone := status.StartWatchdogCheckCounter(ctx, watchdogInterval)
	managementDone := management.Start(ctx, provider, controller, log.Printf)
	log.Printf("browser watchdog started with interval %s", watchdogInterval)
	log.Printf("http server listening on %s", cfg.HTTPAddr)

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
	<-watchdogMetricsDone
	<-managementDone
}
