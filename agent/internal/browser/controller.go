package browser

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const kioskService = "kiosk.service"

// Status is the browser state obtained by one concrete controller.
type Status struct {
	Name        string
	Running     bool
	PID         int
	Version     string
	Executable  string
	CommandLine string
}

// Controller limits browser management to known local implementations.
// Remote commands cannot supply paths, service names, or shell fragments.
type Controller interface {
	Inspect(context.Context) Status
	Restart(context.Context) error
	Reload(context.Context) error
}

// NewController selects a controller from validated local configuration.
func NewController(controller, browserName, service string) (Controller, error) {
	switch controller {
	case "supervisor":
		return SupervisorController{Runtime: NewRuntime(browserName)}, nil
	case "systemd-service":
		if service != kioskService {
			return nil, fmt.Errorf("systemd browser service must be %s", kioskService)
		}
		return NewSystemdServiceController(browserName), nil
	default:
		return nil, fmt.Errorf("unsupported browser controller %q", controller)
	}
}

// SupervisorController preserves the existing Cage/browser-supervisor path.
type SupervisorController struct {
	Runtime Runtime
}

func (c SupervisorController) Inspect(context.Context) Status {
	return runtimeStatus(c.Runtime)
}

func (c SupervisorController) Restart(context.Context) error { return RestartService() }

func (c SupervisorController) Reload(context.Context) error { return ReloadService() }

func runtimeStatus(runtime Runtime) Status {
	return Status{
		Name:        runtime.Name,
		Running:     runtime.IsRunning(),
		PID:         runtime.PID(),
		Version:     runtime.Version(),
		Executable:  runtime.Executable(),
		CommandLine: runtime.CommandLine(),
	}
}

type commandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// SystemdServiceController manages only the fixed Mini-PC kiosk.service.
type SystemdServiceController struct {
	browserName string
	runner      commandRunner
	inspectPID  func(int) (processInfo, bool)
	timeout     time.Duration
	poll        time.Duration
}

// NewSystemdServiceController creates the management-only controller.
func NewSystemdServiceController(browserName string) *SystemdServiceController {
	runtime := NewRuntime(browserName)
	return &SystemdServiceController{
		browserName: runtime.Name,
		runner:      execRunner{},
		inspectPID:  runtime.readProcess,
		timeout:     10 * time.Second,
		poll:        200 * time.Millisecond,
	}
}

func (c *SystemdServiceController) Inspect(ctx context.Context) Status {
	inspectCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	properties, err := c.show(inspectCtx)
	if err != nil {
		return Status{Name: c.browserName}
	}
	pid, err := strconv.Atoi(properties["MainPID"])
	if err != nil || pid <= 0 || properties["LoadState"] != "loaded" || properties["ActiveState"] != "active" || properties["SubState"] != "running" {
		return Status{Name: c.browserName}
	}
	process, ok := c.inspectPID(pid)
	if !ok || !isKioskChromium(process) {
		return Status{Name: c.browserName}
	}
	return Status{
		Name: c.browserName, Running: true, PID: pid,
		Version: readExecutableVersion(process.executable), Executable: process.executable,
		CommandLine: process.commandLine,
	}
}

func (c *SystemdServiceController) Restart(ctx context.Context) error {
	restartCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	if _, err := c.runner.Run(restartCtx, "sudo", "-n", "/usr/bin/systemctl", "restart", kioskService); err != nil {
		return fmt.Errorf("restart %s: %w", kioskService, err)
	}

	ticker := time.NewTicker(c.poll)
	defer ticker.Stop()
	for {
		if c.Inspect(restartCtx).Running {
			return nil
		}
		select {
		case <-restartCtx.Done():
			return fmt.Errorf("restart %s timed out waiting for active chromium kiosk process", kioskService)
		case <-ticker.C:
		}
	}
}

func (c *SystemdServiceController) Reload(context.Context) error {
	return fmt.Errorf("browser reload is not supported by %s controller", kioskService)
}

func (c *SystemdServiceController) show(ctx context.Context) (map[string]string, error) {
	output, err := c.runner.Run(ctx, "/usr/bin/systemctl", "show", kioskService,
		"--property=LoadState", "--property=ActiveState", "--property=SubState", "--property=MainPID")
	if err != nil {
		return nil, err
	}
	properties := make(map[string]string)
	for _, line := range strings.Split(string(output), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			properties[key] = value
		}
	}
	return properties, nil
}

func isKioskChromium(process processInfo) bool {
	commandLine := strings.ToLower(process.commandLine)
	executable := strings.ToLower(filepath.Base(process.executable))
	return (strings.Contains(executable, "chrom") || strings.Contains(commandLine, "chrom")) && strings.Contains(commandLine, "--kiosk")
}
