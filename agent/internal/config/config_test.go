package config

import (
	"os"
	"path/filepath"
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
