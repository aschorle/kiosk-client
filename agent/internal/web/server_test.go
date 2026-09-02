package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
	"github.com/aschorle/kiosk-client/agent/internal/status"
)

type testController struct{}

func (testController) Inspect(context.Context) browser.Status { return browser.Status{} }
func (testController) Restart(context.Context) error          { return nil }
func (testController) Reload(context.Context) error           { return nil }

func TestManagementOnlyServerUsesConfiguredLoopbackPortAndRejectsReboot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "client.conf")
	content := "URL=http://example.test\nBROWSER=chromium\nCLIENT_TYPE=systemd\nBROWSER_CONTROLLER=systemd-service\nBROWSER_SERVICE=kiosk.service\nBROWSER_WATCHDOG=observe\nENABLE_REBOOT=false\nHTTP_ADDR=127.0.0.1:18080\nCONFIG_WRITABLE=false\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(cfg.HTTPAddr, status.NewProvider(cfg, "test", testController{}), testController{})
	if server.addr != "127.0.0.1:18080" {
		t.Fatalf("unexpected management-only API address: %q", server.addr)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/system/reboot", nil)
	response := httptest.NewRecorder()
	server.mux().ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("reboot response = %d, want %d", response.Code, http.StatusForbidden)
	}
}
