// Package management implements the outbound central-management protocol.
package management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	Command       *command       `json:"command"`
	DesiredConfig *desiredConfig `json:"desired_config"`
}

type desiredConfig struct {
	BrowserURL string          `json:"browser_url"`
	Revision   json.RawMessage `json:"revision"`
}

// reportedConfig intentionally contains no authentication material.
type reportedConfig struct {
	BrowserURL      string `json:"browser_url,omitempty"`
	AppliedRevision int    `json:"applied_revision,omitempty"`
	Status          string `json:"status"`
	ErrorRevision   int    `json:"error_revision,omitempty"`
	ErrorCode       string `json:"error_code,omitempty"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type managementState struct {
	EffectiveBrowserURL string `json:"effective_browser_url,omitempty"`
	DesiredRevision     string `json:"desired_revision,omitempty"`
	AppliedRevision     string `json:"applied_revision,omitempty"`
	Status              string `json:"status"`
	ErrorRevision       string `json:"error_revision,omitempty"`
	ErrorCode           string `json:"error_code,omitempty"`
	ErrorMessage        string `json:"error_message,omitempty"`
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
	state, stateErr := loadStateFor(cfg)
	if stateErr != nil {
		state = managementState{Status: "error", ErrorCode: "state_corrupt", ErrorMessage: "management state is invalid"}
	}
	payload := heartbeatPayload(cfg, current, health, acks, state)
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
	if supportsManagedBrowserURL(cfg) && result.DesiredConfig != nil {
		if err := applyDesired(ctx, cfg, controller, state, *result.DesiredConfig); err != nil {
			logf("central browser URL apply failed: %v", err)
		}
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
func heartbeatPayload(cfg config.Config, current status.Status, health status.Health, acks []ack, state managementState) map[string]any {
	payload := map[string]any{
		"client_id": cfg.DeviceID, "name": deviceName(cfg), "client_type": protocolClientType(cfg),
		"agent_version": status.AgentVersion, "browser_running": current.BrowserRunning,
		"health": health.Status, "capabilities": capabilities(cfg), "acks": acks,
	}
	payload["protocol_version"] = 2
	payload["config_capabilities"] = []string{}
	if supportsManagedBrowserURL(cfg) {
		payload["config_capabilities"] = []string{"browser_url"}
	} else {
		state = managementState{Status: "not_supported"}
	}
	payload["reported_config"] = reportState(state)
	return payload
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

func supportsManagedBrowserURL(cfg config.Config) bool {
	// Only the appliance supervisor restarts scripts/start-browser.sh, which
	// reads the managed state. The Mini-PC kiosk.service is intentionally not
	// altered in this phase.
	return cfg.ClientType != "systemd" && cfg.BrowserController == "supervisor"
}

func reportState(state managementState) reportedConfig {
	report := reportedConfig{BrowserURL: state.EffectiveBrowserURL, Status: state.Status, ErrorCode: state.ErrorCode, ErrorMessage: state.ErrorMessage}
	if revision, err := strconv.Atoi(state.AppliedRevision); err == nil && revision > 0 {
		report.AppliedRevision = revision
	}
	if revision, err := strconv.Atoi(state.ErrorRevision); err == nil && revision > 0 {
		report.ErrorRevision = revision
	}
	return report
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

func revision(raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err == nil && number.String() != "" {
		return number.String(), nil
	}
	return "", fmt.Errorf("desired config revision is invalid")
}

func applyDesired(ctx context.Context, _ config.Config, controller browser.Controller, state managementState, desired desiredConfig) error {
	rev, err := revision(desired.Revision)
	if err != nil {
		return err
	}
	// A completed or terminally rejected revision is never retried. A newer
	// revision clears stale errors and is eligible for one apply attempt.
	if rev == state.AppliedRevision || rev == state.ErrorRevision {
		return nil
	}
	path, err := config.ManagementStatePath()
	if err != nil {
		return err
	}
	state.DesiredRevision = rev
	state.ErrorRevision, state.ErrorCode, state.ErrorMessage = "", "", ""
	if err := validateBrowserURL(desired.BrowserURL); err != nil {
		state.Status, state.ErrorRevision, state.ErrorCode, state.ErrorMessage = "error", rev, "invalid_url", err.Error()
		return saveState(path, state)
	}
	// Persist the intended URL before signalling the supervisor: it may start
	// Chromium immediately after receiving the signal and must observe it.
	previousURL := state.EffectiveBrowserURL
	state.Status = "pending"
	state.EffectiveBrowserURL = desired.BrowserURL
	if err := saveState(path, state); err != nil {
		return err
	}
	if desired.BrowserURL != previousURL {
		if err := controller.Restart(ctx); err != nil {
			state.EffectiveBrowserURL = previousURL
			state.Status, state.ErrorRevision, state.ErrorCode, state.ErrorMessage = "error", rev, "restart_failed", "browser restart failed"
			if saveErr := saveState(path, state); saveErr != nil {
				return fmt.Errorf("restart: %v; save error: %w", err, saveErr)
			}
			return fmt.Errorf("browser restart: %w", err)
		}
	}
	state.AppliedRevision, state.Status = rev, "synced"
	return saveState(path, state)
}

func validateBrowserURL(value string) error {
	if value == "" || len(value) > 2048 {
		return fmt.Errorf("browser URL must be 1 to 2048 characters")
	}
	for _, r := range value {
		if r == '\\' || r <= 0x1f || r == 0x7f || unicode.IsSpace(r) {
			return fmt.Errorf("browser URL contains invalid whitespace or control characters")
		}
	}
	u, err := url.ParseRequestURI(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return fmt.Errorf("browser URL must be an absolute http or https URL without userinfo")
	}
	if u.Hostname() == "" {
		return fmt.Errorf("browser URL host is required")
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return fmt.Errorf("browser URL port is invalid")
		}
	}
	return nil
}

func loadStateFor(config.Config) (managementState, error) {
	path, err := config.ManagementStatePath()
	if err != nil {
		return managementState{Status: "not_initialized"}, err
	}
	return loadState(path)
}

func loadState(path string) (managementState, error) {
	contents, err := os.ReadFile(filepath.Clean(path))
	if os.IsNotExist(err) {
		return managementState{Status: "not_initialized"}, nil
	}
	if err != nil {
		return managementState{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(contents, &fields); err != nil {
		return managementState{}, err
	}
	allowed := map[string]bool{"effective_browser_url": true, "desired_revision": true, "applied_revision": true, "status": true, "error_revision": true, "error_code": true, "error_message": true}
	for key := range fields {
		if !allowed[key] {
			return managementState{}, fmt.Errorf("unknown management state field %q", key)
		}
	}
	var state managementState
	if err := json.Unmarshal(contents, &state); err != nil {
		return managementState{}, err
	}
	if state.Status == "" {
		return managementState{}, fmt.Errorf("management state status is missing")
	}
	return state, nil
}

func saveState(path string, state managementState) error {
	contents, err := json.Marshal(state)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".management-state-*")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if err := file.Chmod(0600); err == nil {
		_, err = file.Write(append(contents, '\n'))
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
