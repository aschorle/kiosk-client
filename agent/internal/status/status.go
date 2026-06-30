package status

import (
	"net"
	"os"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
)

// Status is the public runtime status returned by the local HTTP API.
type Status struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	URL      string `json:"url"`
	Browser  string `json:"browser"`
	Version  string `json:"version"`
}

// Provider builds status responses from static configuration and local system
// information.
type Provider struct {
	config  config.Config
	version string
}

// NewProvider creates a status provider for the current agent process.
func NewProvider(cfg config.Config, version string) Provider {
	return Provider{
		config:  cfg,
		version: version,
	}
}

// Current returns the current kiosk-client status.
func (p Provider) Current() Status {
	runtime := browser.NewRuntime(p.config.Browser)

	return Status{
		Hostname: hostname(),
		IP:       primaryIP(),
		URL:      p.config.URL,
		Browser:  runtime.Name,
		Version:  p.version,
	}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}

	return name
}

func primaryIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}

		return ip.String()
	}

	return ""
}
