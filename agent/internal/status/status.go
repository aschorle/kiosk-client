package status

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/aschorle/kiosk-client/agent/internal/browser"
	"github.com/aschorle/kiosk-client/agent/internal/config"
)

const (
	agentVersion = "0.10.4"
	clockTicks   = 100
)

var (
	BuildTime         = "unknown"
	GitCommit         = "unknown"
	agentStartedAt    = time.Now().UTC()
	httpRequestsTotal atomic.Uint64
	watchdogChecks    atomic.Uint64
)

// Status is the public runtime status returned by the local HTTP API.
type Status struct {
	Hostname              string                 `json:"hostname"`
	IP                    string                 `json:"ip"`
	URL                   string                 `json:"url"`
	Browser               string                 `json:"browser"`
	Version               string                 `json:"version"`
	BrowserRunning        bool                   `json:"browser_running"`
	BrowserPID            int                    `json:"browser_pid"`
	BrowserVersion        string                 `json:"browser_version"`
	BrowserPath           string                 `json:"browser_path"`
	BrowserCmdline        string                 `json:"browser_cmdline"`
	BrowserRestartCount   uint64                 `json:"browser_restart_count"`
	BrowserLastRestart    string                 `json:"browser_last_restart"`
	BrowserWatchdogState  string                 `json:"browser_watchdog_state"`
	BrowserRestartHistory []browser.RestartEvent `json:"browser_restart_history"`
	Uptime                string                 `json:"uptime"`
	Kernel                string                 `json:"kernel"`
	DebianVersion         string                 `json:"debian_version"`
	Architecture          string                 `json:"architecture"`
	CPUModel              string                 `json:"cpu_model"`
	MemoryTotal           uint64                 `json:"memory_total"`
	MemoryAvailable       uint64                 `json:"memory_available"`
	DiskTotal             uint64                 `json:"disk_total"`
	DiskAvailable         uint64                 `json:"disk_available"`
	LoadAverage           string                 `json:"load_average"`
}

// Info is the public static and runtime information returned by /api/info.
type Info struct {
	AgentVersion string `json:"agent_version"`
	GoVersion    string `json:"go_version"`
	Hostname     string `json:"hostname"`
	Architecture string `json:"architecture"`
	Kernel       string `json:"kernel"`
	BuildTime    string `json:"build_time"`
	GitCommit    string `json:"git_commit"`
	Board        string `json:"board"`
	OSName       string `json:"os_name"`
	OSVersion    string `json:"os_version"`
}

// Health is the summarized system health returned by /api/health.
type Health struct {
	Status string `json:"status"`
}

// Metrics contains runtime counters and memory statistics returned by /api/metrics.
type Metrics struct {
	AgentUptimeSeconds   uint64 `json:"agent_uptime_seconds"`
	BrowserUptimeSeconds uint64 `json:"browser_uptime_seconds"`
	WatchdogChecks       uint64 `json:"watchdog_checks"`
	BrowserRestartCount  uint64 `json:"browser_restart_count"`
	HTTPRequestsTotal    uint64 `json:"http_requests_total"`
	Goroutines           int    `json:"goroutines"`
	MemoryAllocBytes     uint64 `json:"memory_alloc_bytes"`
	MemorySysBytes       uint64 `json:"memory_sys_bytes"`
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

// SetAgentStartTime sets the timestamp used for agent uptime metrics.
func SetAgentStartTime(startedAt time.Time) {
	agentStartedAt = startedAt.UTC()
}

// IncrementHTTPRequest increments the HTTP request counter.
func IncrementHTTPRequest() {
	httpRequestsTotal.Add(1)
}

// IncrementWatchdogCheck increments the watchdog check counter.
func IncrementWatchdogCheck() {
	watchdogChecks.Add(1)
}

// StartWatchdogCheckCounter counts expected watchdog checks until ctx is done.
func StartWatchdogCheckCounter(ctx context.Context, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})

	if interval <= 0 {
		interval = 30 * time.Second
	}

	go func() {
		defer close(done)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				IncrementWatchdogCheck()
			}
		}
	}()

	return done
}

// Current returns the current kiosk-client status.
func (p Provider) Current() Status {
	cfg := p.currentConfig()
	browserRuntime := browser.NewRuntime(cfg.Browser)

	return Status{
		Hostname:              hostname(),
		IP:                    primaryIP(),
		URL:                   cfg.URL,
		Browser:               browserRuntime.Name,
		Version:               p.version,
		BrowserRunning:        browserRuntime.IsRunning(),
		BrowserPID:            browserRuntime.PID(),
		BrowserVersion:        browserRuntime.Version(),
		BrowserPath:           browserRuntime.Executable(),
		BrowserCmdline:        browserRuntime.CommandLine(),
		BrowserRestartCount:   browser.RestartCount(),
		BrowserLastRestart:    browser.LastRestart(),
		BrowserWatchdogState:  browser.WatchdogState(),
		BrowserRestartHistory: browser.RestartHistory(),
		Uptime:                uptime(),
		Kernel:                kernel(),
		DebianVersion:         debianVersion(),
		Architecture:          architecture(),
		CPUModel:              cpuModel(),
		MemoryTotal:           memoryTotal(),
		MemoryAvailable:       memoryAvailable(),
		DiskTotal:             diskTotal(),
		DiskAvailable:         diskAvailable(),
		LoadAverage:           loadAverage(),
	}
}

// Metrics returns agent, browser, HTTP, watchdog, and Go runtime metrics.
func (p Provider) Metrics() Metrics {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)

	return Metrics{
		AgentUptimeSeconds:   agentUptimeSeconds(),
		BrowserUptimeSeconds: browserUptimeSeconds(),
		WatchdogChecks:       watchdogChecks.Load(),
		BrowserRestartCount:  browser.RestartCount(),
		HTTPRequestsTotal:    httpRequestsTotal.Load(),
		Goroutines:           runtime.NumGoroutine(),
		MemoryAllocBytes:     memory.Alloc,
		MemorySysBytes:       memory.Sys,
	}
}

// Health returns the summarized system health.
func (p Provider) Health() Health {
	cfg := p.currentConfig()
	browserRuntime := browser.NewRuntime(cfg.Browser)
	if !browserRuntime.IsRunning() {
		return Health{Status: "error"}
	}

	if browser.WatchdogState() == "limited" {
		return Health{Status: "degraded"}
	}

	if browser.WatchdogState() == "healthy" {
		return Health{Status: "healthy"}
	}

	return Health{Status: "degraded"}
}

func (p Provider) currentConfig() config.Config {
	cfg, err := config.Current()
	if err != nil {
		return p.config
	}

	return cfg
}

// Info returns general agent, OS, and board information.
func (p Provider) Info() Info {
	osName, osVersion := osRelease()

	return Info{
		AgentVersion: agentVersion,
		GoVersion:    runtime.Version(),
		Hostname:     hostname(),
		Architecture: architecture(),
		Kernel:       kernel(),
		BuildTime:    valueOrUnknown(BuildTime),
		GitCommit:    valueOrUnknown(GitCommit),
		Board:        board(),
		OSName:       osName,
		OSVersion:    osVersion,
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

func agentUptimeSeconds() uint64 {
	if agentStartedAt.IsZero() {
		return 0
	}

	uptime := time.Since(agentStartedAt)
	if uptime < 0 {
		return 0
	}

	return uint64(uptime.Seconds())
}

func browserUptimeSeconds() uint64 {
	pid := browser.PID()
	if pid <= 0 {
		return 0
	}

	processStart, ok := processStartSeconds(pid)
	if !ok {
		return 0
	}

	systemUptime, ok := systemUptimeSeconds()
	if !ok || systemUptime < processStart {
		return 0
	}

	return uint64(systemUptime - processStart)
}

func processStartSeconds(pid int) (float64, bool) {
	content, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return 0, false
	}

	line := string(content)
	commandEnd := strings.LastIndex(line, ")")
	if commandEnd < 0 || commandEnd+2 >= len(line) {
		return 0, false
	}

	fields := strings.Fields(line[commandEnd+2:])
	if len(fields) <= 19 {
		return 0, false
	}

	startTicks, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		return 0, false
	}

	return float64(startTicks) / clockTicks, true
}

func systemUptimeSeconds() (float64, bool) {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, false
	}

	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return 0, false
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, false
	}

	return seconds, true
}

func board() string {
	paths := []string{
		"/proc/device-tree/model",
		"/sys/firmware/devicetree/base/model",
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		value := strings.Trim(strings.TrimSpace(string(content)), "\x00")
		if value != "" {
			return value
		}
	}

	return "unknown"
}

func osRelease() (string, string) {
	fields := readKeyValueFile("/etc/os-release")
	name := fields["PRETTY_NAME"]
	version := fields["VERSION_ID"]

	if name == "" {
		name = fields["NAME"]
	}

	if version == "" {
		version = debianVersion()
	}

	return valueOrUnknown(name), valueOrUnknown(version)
}

func readKeyValueFile(path string) map[string]string {
	result := make(map[string]string)

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		result[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
	}

	return result
}

func valueOrUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}

	return value
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
