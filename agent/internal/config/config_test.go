package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetForTest() {
	mu.Lock()
	defer mu.Unlock()
	currentConfig = Config{}
	currentError = nil
	currentPath = ""
	loaded = false
}

func TestDefaultConfigurationRemainsApplianceCompatible(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
	path := filepath.Join(t.TempDir(), "missing.conf")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ClientType != defaultClientType || cfg.BrowserController != defaultBrowserController || cfg.WatchdogMode != defaultWatchdogMode || !cfg.EnableReboot || cfg.HTTPAddr != defaultHTTPAddr || !cfg.ConfigWritable {
		t.Fatalf("unexpected appliance defaults: %#v", cfg)
	}
}

func TestSystemdManagementOnlyConfiguration(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
	path := filepath.Join(t.TempDir(), "client.conf")
	content := "URL=http://example.test/\nBROWSER=chromium\nCLIENT_TYPE=systemd\nBROWSER_CONTROLLER=systemd-service\nBROWSER_SERVICE=kiosk.service\nBROWSER_WATCHDOG=observe\nENABLE_REBOOT=false\nHTTP_ADDR=127.0.0.1:18080\nCONFIG_WRITABLE=false\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ClientType != systemdClientType || cfg.BrowserService != systemdBrowserService || cfg.EnableReboot || cfg.ConfigWritable || cfg.HTTPAddr != "127.0.0.1:18080" {
		t.Fatalf("unexpected management-only config: %#v", cfg)
	}
}

func TestDisplayConfigurationSurvivesDashboardUpdateButIsNotExposed(t *testing.T) {
	resetForTest()
	t.Cleanup(resetForTest)
	path := filepath.Join(t.TempDir(), "client.conf")
	content := "URL=http://example.test/\nBROWSER=chromium\nDISPLAY_OUTPUT=HDMI-A-1\nDISPLAY_MODE=1680x1050@59.883Hz\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DisplayOutput != "HDMI-A-1" || cfg.DisplayMode != "1680x1050@59.883Hz" {
		t.Fatalf("display configuration not loaded: %#v", cfg)
	}
	publicConfig, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(publicConfig), "DISPLAY_OUTPUT") || strings.Contains(string(publicConfig), "display_mode") {
		t.Fatalf("local display configuration unexpectedly exposed in API JSON: %s", publicConfig)
	}

	if err := Update(Config{URL: "http://updated.example/", DeviceID: "display02", Browser: "chromium"}); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "DISPLAY_OUTPUT=HDMI-A-1\n") || !strings.Contains(string(updated), "DISPLAY_MODE=1680x1050@59.883Hz\n") {
		t.Fatalf("dashboard update lost local display configuration:\n%s", updated)
	}
	if !strings.Contains(string(updated), "updated.example") {
		t.Fatalf("dashboard update was not written:\n%s", updated)
	}
}

func TestSystemdConfigurationRejectsUnsafeValues(t *testing.T) {
	base := Config{URL: "http://example.test", Browser: "chromium", ClientType: systemdClientType, BrowserController: systemdBrowserController, BrowserService: "other.service", WatchdogMode: managementOnlyWatchdogMode, HTTPAddr: "127.0.0.1:18080"}
	if _, err := Validate(base); err == nil {
		t.Fatal("expected arbitrary service to be rejected")
	}
	base.BrowserService = systemdBrowserService
	base.EnableReboot = true
	if _, err := Validate(base); err == nil {
		t.Fatal("expected systemd reboot to be rejected")
	}
	base.EnableReboot = false
	base.HTTPAddr = "0.0.0.0:18080"
	if _, err := Validate(base); err == nil {
		t.Fatal("expected non-loopback address to be rejected")
	}
}
