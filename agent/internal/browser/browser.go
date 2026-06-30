package browser

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultName    = "chromium"
	dpkgStatus    = "/var/lib/dpkg/status"
	serviceName    = "kiosk-browser.service"
	systemctl      = "systemctl"
	userSystemctl  = "--user"
	reloadProperty = "CanReload"
)

// ErrReloadNotSupported is returned when systemd reports that the browser
// service has no reload action.
var ErrReloadNotSupported = errors.New("reload not supported")

var watchdogMetrics struct {
	sync.Mutex
	restartCount uint64
	lastRestart  string
}

// Runtime describes the configured browser executable.
//
// The agent only reads process information. It does not start, stop, restart,
// signal, or otherwise modify Chromium.
type Runtime struct {
	Name string
}

// NewRuntime returns a browser runtime descriptor without starting a process.
func NewRuntime(name string) Runtime {
	return Runtime{Name: name}
}

// IsRunning reports whether the default Chromium process is currently visible.
func IsRunning() bool {
	return NewRuntime(defaultName).IsRunning()
}

// PID returns the process id of the default Chromium process.
func PID() int {
	return NewRuntime(defaultName).PID()
}

// Version returns the detected version of the default Chromium process.
func Version() string {
	return NewRuntime(defaultName).Version()
}

// Executable returns the executable path of the default Chromium process.
func Executable() string {
	return NewRuntime(defaultName).Executable()
}

// CommandLine returns the command line of the default Chromium process.
func CommandLine() string {
	return NewRuntime(defaultName).CommandLine()
}

// Restart restarts the browser through the systemd user manager.
func Restart() error {
	return RestartService()
}

// RestartService restarts the browser through the systemd user manager.
func RestartService() error {
	return runUserSystemctl("restart", serviceName)
}

// ReloadService reloads the browser service when systemd reports support for it.
func ReloadService() error {
	canReload, err := serviceCanReload()
	if err != nil {
		return err
	}

	if !canReload {
		return ErrReloadNotSupported
	}

	return runUserSystemctl("reload", serviceName)
}

// RestartCount returns the number of watchdog-triggered browser restarts.
func RestartCount() uint64 {
	watchdogMetrics.Lock()
	defer watchdogMetrics.Unlock()

	return watchdogMetrics.restartCount
}

// LastRestart returns the UTC timestamp of the last watchdog restart.
func LastRestart() string {
	watchdogMetrics.Lock()
	defer watchdogMetrics.Unlock()

	return watchdogMetrics.lastRestart
}

// StartWatchdog starts a background browser watchdog worker.
func StartWatchdog(ctx context.Context, interval time.Duration, logf func(string, ...interface{})) <-chan struct{} {
	done := make(chan struct{})

	if interval <= 0 {
		interval = 30 * time.Second
	}

	if logf == nil {
		logf = log.Printf
	}

	go func() {
		defer close(done)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logf("browser watchdog stopped")
				return
			case <-ticker.C:
				if IsRunning() {
					continue
				}

				if err := Restart(); err != nil {
					logf("browser watchdog restart failed: %v", err)
					continue
				}

				restartedAt := time.Now().UTC().Format(time.RFC3339)
				recordRestart(restartedAt)
				logf("browser watchdog restarted kiosk-browser.service at %s", restartedAt)
			}
		}
	}()

	return done
}

// IsRunning reports whether a matching browser process is currently visible.
func (r Runtime) IsRunning() bool {
	return r.PID() > 0
}

// PID returns the process id of the first matching browser process.
func (r Runtime) PID() int {
	process, ok := r.findProcess()
	if !ok {
		return 0
	}

	return process.pid
}

// Version returns the detected browser version for the running executable.
func (r Runtime) Version() string {
	if !r.IsRunning() {
		return ""
	}

	for _, packageName := range r.packageNames() {
		if version := readPackageVersion(dpkgStatus, packageName); version != "" {
			return version
		}
	}

	return readExecutableVersion(r.Executable())
}

// Executable returns the executable path of the running browser process.
func (r Runtime) Executable() string {
	process, ok := r.findProcess()
	if !ok {
		return ""
	}

	return process.executable
}

// CommandLine returns the command line of the running browser process.
func (r Runtime) CommandLine() string {
	process, ok := r.findProcess()
	if !ok {
		return ""
	}

	return process.commandLine
}

type processInfo struct {
	pid         int
	executable  string
	commandLine string
}

func (r Runtime) findProcess() (processInfo, bool) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return processInfo{}, false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		process, ok := r.readProcess(pid)
		if !ok {
			continue
		}

		if r.matches(process) {
			return process, true
		}
	}

	return processInfo{}, false
}

func (r Runtime) readProcess(pid int) (processInfo, bool) {
	procDir := filepath.Join("/proc", strconv.Itoa(pid))

	executable, err := os.Readlink(filepath.Join(procDir, "exe"))
	if err != nil {
		executable = ""
	}

	commandLine, err := readCommandLine(filepath.Join(procDir, "cmdline"))
	if err != nil || commandLine == "" {
		return processInfo{}, false
	}

	return processInfo{
		pid:         pid,
		executable:  executable,
		commandLine: commandLine,
	}, true
}

func (r Runtime) matches(process processInfo) bool {
	name := r.browserName()
	executableName := strings.ToLower(filepath.Base(process.executable))
	commandLine := strings.ToLower(process.commandLine)

	if executableName == name || executableName == "chromium-browser" {
		return true
	}

	return strings.Contains(commandLine, name) || strings.Contains(commandLine, "chromium-browser")
}

func (r Runtime) browserName() string {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return defaultName
	}

	return strings.ToLower(filepath.Base(name))
}

func (r Runtime) packageNames() []string {
	name := r.browserName()
	names := []string{name}

	if name != "chromium" {
		names = append(names, "chromium")
	}

	if name != "chromium-x11" {
		names = append(names, "chromium-x11")
	}

	if name != "chromium-browser" {
		names = append(names, "chromium-browser")
	}

	return names
}

func readCommandLine(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content = bytes.TrimRight(content, "\x00")
	parts := bytes.Split(content, []byte{0})
	args := make([]string, 0, len(parts))

	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		args = append(args, string(part))
	}

	return strings.Join(args, " "), nil
}

func readExecutableVersion(executable string) string {
	if executable == "" {
		return ""
	}

	output, err := exec.Command(executable, "--version").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

func serviceCanReload() (bool, error) {
	output, err := exec.Command(systemctl, userSystemctl, "show", serviceName, "-p", reloadProperty, "--value").Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) == "yes", nil
}

func runUserSystemctl(args ...string) error {
	commandArgs := append([]string{userSystemctl}, args...)
	return exec.Command(systemctl, commandArgs...).Run()
}

func recordRestart(restartedAt string) {
	watchdogMetrics.Lock()
	defer watchdogMetrics.Unlock()

	watchdogMetrics.restartCount++
	watchdogMetrics.lastRestart = restartedAt
}

func readPackageVersion(path string, packageName string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	paragraphs := strings.Split(string(content), "\n\n")
	for _, paragraph := range paragraphs {
		fields := parsePackageParagraph(paragraph)
		if fields["Package"] == packageName {
			return fields["Version"]
		}
	}

	return ""
}

func parsePackageParagraph(paragraph string) map[string]string {
	fields := make(map[string]string)
	lines := strings.Split(paragraph, "\n")

	for _, line := range lines {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return fields
}
