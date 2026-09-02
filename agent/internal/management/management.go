// Package management implements the outbound central-management protocol.
package management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
	"github.com/aschorle/kiosk-client/agent/internal/status"
	"github.com/aschorle/kiosk-client/agent/internal/web"
)

const interval = 20 * time.Second

type ack struct {
	CommandID string `json:"command_id"`
	Result    string `json:"result"`
	Message   string `json:"message,omitempty"`
}

type command struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type heartbeatResponse struct {
	Command *command `json:"command"`
}

// Start starts no worker when central management is not configured.
func Start(ctx context.Context, provider status.Provider, controller browser.Controller, logf func(string, ...any)) <-chan struct{} {
	done := make(chan struct{})
	cfg, err := config.Current()
	if err != nil || cfg.ServerURL == "" {
		close(done)
		return done
	}
	if strings.TrimSpace(cfg.DeviceID) == "" || strings.TrimSpace(cfg.AuthToken) == "" {
		logf("central management disabled: DEVICE_ID and AUTH_TOKEN are required")
		close(done)
		return done
	}
	go func() {
		defer close(done)
		client := &http.Client{Timeout: 10 * time.Second}
		for {
			if err := heartbeat(ctx, client, provider, controller, []ack{}, logf); err != nil && ctx.Err() == nil {
				logf("central management heartbeat failed: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
		}
	}()
	return done
}

func heartbeat(ctx context.Context, client *http.Client, provider status.Provider, controller browser.Controller, acks []ack, logf func(string, ...any)) error {
	cfg, err := config.Current()
	if err != nil {
		return err
	}
	current, health := provider.Current(), provider.Health()
	payload := heartbeatPayload(cfg, current, health, acks)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.ServerURL+"/api/clients/heartbeat", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", response.Status)
	}
	var result heartbeatResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return err
	}
	if result.Command == nil {
		return nil
	}
	completed := execute(ctx, *result.Command, controller, cfg)
	// Persist the acknowledgement immediately. A reboot starts only after this
	// acknowledgement heartbeat has completed successfully.
	if err := heartbeat(ctx, client, provider, controller, []ack{completed}, logf); err != nil {
		return err
	}
	if result.Command.Action == "reboot" && cfg.EnableReboot && completed.Result == "success" {
		time.AfterFunc(200*time.Millisecond, func() {
			if err := web.RebootSystem(); err != nil {
				logf("central reboot failed: %v", err)
			}
		})
	}
	return nil
}

// heartbeatPayload keeps the externally visible protocol value independent of
// the internal profile/controller selection. Legacy appliance installations
// therefore continue to identify themselves as kiosk-client.
func heartbeatPayload(cfg config.Config, current status.Status, health status.Health, acks []ack) map[string]any {
	return map[string]any{
		"client_id": cfg.DeviceID, "name": deviceName(cfg), "client_type": protocolClientType(cfg),
		"agent_version": status.AgentVersion, "browser_running": current.BrowserRunning,
		"health": health.Status, "capabilities": capabilities(cfg), "acks": acks,
	}
}

func protocolClientType(cfg config.Config) string {
	switch cfg.ClientType {
	case "", "appliance":
		return "kiosk-client"
	case "systemd":
		return "mini-kiosk"
	default:
		return cfg.ClientType
	}
}

func capabilities(cfg config.Config) []string {
	capabilities := []string{"restart_browser"}
	if cfg.EnableReboot {
		capabilities = append(capabilities, "reboot")
	}
	return capabilities
}

func deviceName(cfg config.Config) string {
	if strings.TrimSpace(cfg.DeviceName) != "" {
		return cfg.DeviceName
	}
	return cfg.DeviceID
}

func execute(ctx context.Context, item command, controller browser.Controller, cfg config.Config) ack {
	result := ack{CommandID: item.ID, Result: "failed"}
	switch item.Action {
	case "restart_browser":
		if err := controller.Restart(ctx); err != nil {
			result.Message = "browser restart failed"
		} else {
			result.Result, result.Message = "success", "browser restart requested"
		}
	case "reboot":
		if !cfg.EnableReboot {
			result.Message = "unsupported command"
			break
		}
		// The fixed reboot is scheduled only after the acknowledgement heartbeat.
		result.Result, result.Message = "success", "reboot requested"
	default:
		result.Message = "unsupported command"
	}
	return result
}
