package status

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
)

// Status is the public runtime status returned by the local HTTP API.
type Status struct {
	Hostname        string `json:"hostname"`
	IP              string `json:"ip"`
	URL             string `json:"url"`
	Browser         string `json:"browser"`
	Version         string `json:"version"`
	BrowserRunning  bool   `json:"browser_running"`
	BrowserPID      int    `json:"browser_pid"`
	BrowserVersion  string `json:"browser_version"`
	BrowserPath     string `json:"browser_path"`
	BrowserCmdline  string `json:"browser_cmdline"`
	Uptime          string `json:"uptime"`
	Kernel          string `json:"kernel"`
	DebianVersion   string `json:"debian_version"`
	Architecture    string `json:"architecture"`
	CPUModel        string `json:"cpu_model"`
	MemoryTotal     uint64 `json:"memory_total"`
	MemoryAvailable uint64 `json:"memory_available"`
	DiskTotal       uint64 `json:"disk_total"`
	DiskAvailable   uint64 `json:"disk_available"`
	LoadAverage     string `json:"load_average"`
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
	browserRuntime := browser.NewRuntime(p.config.Browser)

	return Status{
		Hostname:        hostname(),
		IP:              primaryIP(),
		URL:             p.config.URL,
		Browser:         browserRuntime.Name,
		Version:         p.version,
		BrowserRunning:  browserRuntime.IsRunning(),
		BrowserPID:      browserRuntime.PID(),
		BrowserVersion:  browserRuntime.Version(),
		BrowserPath:     browserRuntime.Executable(),
		BrowserCmdline:  browserRuntime.CommandLine(),
		Uptime:          uptime(),
		Kernel:          kernel(),
		DebianVersion:   debianVersion(),
		Architecture:    architecture(),
		CPUModel:        cpuModel(),
		MemoryTotal:     memoryTotal(),
		MemoryAvailable: memoryAvailable(),
		DiskTotal:       diskTotal(),
		DiskAvailable:   diskAvailable(),
		LoadAverage:     loadAverage(),
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

func uptime() string {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return ""
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%.0f", seconds)
}

func kernel() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return ""
	}

	return charsToString(uts.Release[:])
}

func debianVersion() string {
	content, err := os.ReadFile("/etc/debian_version")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(content))
}

func architecture() string {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err == nil {
		return charsToString(uts.Machine[:])
	}

	return runtime.GOARCH
}

func cpuModel() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		key = strings.TrimSpace(strings.ToLower(key))
		if key == "model name" || key == "hardware" || key == "processor" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func memoryTotal() uint64 {
	return memoryValue("MemTotal")
}

func memoryAvailable() uint64 {
	return memoryValue("MemAvailable")
}

func memoryValue(name string) uint64 {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, ":")
		if !ok || key != name {
			continue
		}

		fields := strings.Fields(value)
		if len(fields) == 0 {
			return 0
		}

		kilobytes, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return 0
		}

		return kilobytes * 1024
	}

	return 0
}

func diskTotal() uint64 {
	total, _ := diskSpace("/")
	return total
}

func diskAvailable() uint64 {
	_, available := diskSpace("/")
	return available
}

func diskSpace(path string) (uint64, uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0
	}

	blockSize := uint64(stat.Bsize)
	return stat.Blocks * blockSize, stat.Bavail * blockSize
}

func loadAverage() string {
	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(content))
	if len(fields) < 3 {
		return ""
	}

	return strings.Join(fields[:3], " ")
}

func charsToString(chars []int8) string {
	buffer := make([]byte, 0, len(chars))
	for _, char := range chars {
		if char == 0 {
			break
		}

		buffer = append(buffer, byte(char))
	}

	return string(buffer)
}
