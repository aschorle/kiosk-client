package management

import (
	"context"
	"encoding/json"
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
	payload := heartbeatPayload(cfg, status.Status{}, status.Health{Status: "healthy"}, []ack{}, managementState{Status: "not_initialized"})
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

func TestManagedBrowserURLCapabilityOnlyForSupervisor(t *testing.T) {
	if !supportsManagedBrowserURL(config.Config{ClientType: "appliance", BrowserController: "supervisor"}) {
		t.Fatal("supervisor must support managed browser URLs")
	}
	if supportsManagedBrowserURL(config.Config{ClientType: "systemd", BrowserController: "systemd-service"}) {
		t.Fatal("systemd client must not claim managed browser URLs")
	}
}

func TestValidateBrowserURL(t *testing.T) {
	for _, value := range []string{"http://example.test/path?q=x#top", "https://127.0.0.1:8443/", "https://[::1]/"} {
		if err := validateBrowserURL(value); err != nil {
			t.Fatalf("valid URL %q: %v", value, err)
		}
	}
	for _, value := range []string{"", "ftp://example.test", "http://user@example.test", "http://example.test\\x", "http://example.test:99999", "http://example.test/\u00a0"} {
		if err := validateBrowserURL(value); err == nil {
			t.Fatalf("invalid URL accepted: %q", value)
		}
	}
}

func TestStateRoundTripAndRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "management-state.json")
	want := managementState{EffectiveBrowserURL: "https://example.test", DesiredRevision: "2", AppliedRevision: "2", Status: "synced"}
	if err := saveState(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := loadState(path)
	if err != nil || got != want {
		t.Fatalf("state = %#v, %v", got, err)
	}
	if err := os.WriteFile(path, []byte(`{"unexpected":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadState(path); err == nil {
		t.Fatal("unknown state field accepted")
	}
}

func TestReportedAppliedRevisionIsJSONNumber(t *testing.T) {
	payload := heartbeatPayload(config.Config{ClientType: "appliance", BrowserController: "supervisor"}, status.Status{}, status.Health{Status: "healthy"}, nil, managementState{EffectiveBrowserURL: "https://example.test", AppliedRevision: "1", Status: "synced"})
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	reported, ok := decoded["reported_config"].(map[string]any)
	if !ok {
		t.Fatalf("reported_config has type %T", decoded["reported_config"])
	}
	if got, ok := reported["applied_revision"].(float64); !ok || got != 1 {
		t.Fatalf("applied_revision = %#v (%T), want JSON number 1", reported["applied_revision"], reported["applied_revision"])
	}
}

func TestAlreadyAppliedRevisionDoesNotRestart(t *testing.T) {
	controller := &fakeController{}
	result := applyDesired(context.Background(), config.Config{}, controller, managementState{EffectiveBrowserURL: "https://example.test", AppliedRevision: "1", Status: "synced"}, desiredConfig{BrowserURL: "https://other.example", Revision: json.RawMessage(`1`)})
	if result != nil {
		t.Fatal(result)
	}
	if controller.restarts != 0 {
		t.Fatalf("already applied revision restarted browser %d times", controller.restarts)
	}
}
