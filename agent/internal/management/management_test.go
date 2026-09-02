package management

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
	"github.com/aschorle/kiosk-client/agent/internal/status"
)

type fakeController struct {
	restarts int
}

func (c *fakeController) Inspect(context.Context) browser.Status { return browser.Status{} }
func (c *fakeController) Restart(context.Context) error {
	c.restarts++
	return nil
}
func (c *fakeController) Reload(context.Context) error { return nil }

func TestCapabilitiesGateReboot(t *testing.T) {
	if got := capabilities(config.Config{EnableReboot: false}); len(got) != 1 || got[0] != "restart_browser" {
		t.Fatalf("unexpected disabled capabilities: %#v", got)
	}
	if got := capabilities(config.Config{EnableReboot: true}); len(got) != 2 || got[1] != "reboot" {
		t.Fatalf("unexpected appliance capabilities: %#v", got)
	}
}

func TestLegacyConfigurationHeartbeatKeepsKioskClientType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "client.conf")
	legacy := "URL=http://example.test\nBROWSER=chromium\n"
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	payload := heartbeatPayload(cfg, status.Status{}, status.Health{Status: "healthy"}, []ack{})
	if got := payload["client_type"]; got != "kiosk-client" {
		t.Fatalf("legacy heartbeat client_type = %v, want kiosk-client", got)
	}
	if got := payload["capabilities"]; len(got.([]string)) != 2 {
		t.Fatalf("legacy heartbeat capabilities = %#v, want restart_browser and reboot", got)
	}
}

func TestSystemdRebootIsRejectedWithoutAction(t *testing.T) {
	controller := &fakeController{}
	result := execute(context.Background(), command{ID: "test", Action: "reboot"}, controller, config.Config{EnableReboot: false})
	if result.Result != "failed" || result.Message != "unsupported command" || controller.restarts != 0 {
		t.Fatalf("unexpected reboot result: %#v restarts=%d", result, controller.restarts)
	}
}

func TestRestartUsesController(t *testing.T) {
	controller := &fakeController{}
	result := execute(context.Background(), command{ID: "test", Action: "restart_browser"}, controller, config.Config{})
	if result.Result != "success" || controller.restarts != 1 {
		t.Fatalf("unexpected restart result: %#v restarts=%d", result, controller.restarts)
	}
}
