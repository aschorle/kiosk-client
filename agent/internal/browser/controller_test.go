package browser

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeRunner struct {
	showOutput string
	restartErr error
	calls      []string
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, name+" "+strings.Join(args, " "))
	if name == "sudo" {
		return nil, r.restartErr
	}
	return []byte(r.showOutput), nil
}

func activeService() string {
	return "LoadState=loaded\nActiveState=active\nSubState=running\nMainPID=42\n"
}

func kioskProcess(pid int) (processInfo, bool) {
	if pid != 42 {
		return processInfo{}, false
	}
	return processInfo{pid: pid, executable: "/usr/bin/chromium", commandLine: "/usr/bin/chromium --kiosk http://example.test"}, true
}

func TestSystemdControllerReportsOnlyActiveKioskService(t *testing.T) {
	runner := &fakeRunner{showOutput: activeService()}
	controller := NewSystemdServiceController("chromium")
	controller.runner, controller.inspectPID = runner, kioskProcess
	if state := controller.Inspect(context.Background()); !state.Running || state.PID != 42 {
		t.Fatalf("unexpected state: %#v", state)
	}
	runner.showOutput = "LoadState=loaded\nActiveState=inactive\nSubState=dead\nMainPID=0\n"
	if state := controller.Inspect(context.Background()); state.Running {
		t.Fatalf("inactive unit reported as running: %#v", state)
	}
}

func TestNewControllerKeepsSupervisorForAppliance(t *testing.T) {
	controller, err := NewController("supervisor", "chromium", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := controller.(SupervisorController); !ok {
		t.Fatalf("appliance controller type = %T, want SupervisorController", controller)
	}
}

func TestSystemdControllerRestartUsesOnlyFixedService(t *testing.T) {
	runner := &fakeRunner{showOutput: activeService()}
	controller := NewSystemdServiceController("chromium")
	controller.runner, controller.inspectPID = runner, kioskProcess
	controller.timeout, controller.poll = 20*time.Millisecond, time.Millisecond
	if err := controller.Restart(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) < 2 || runner.calls[0] != "sudo -n /usr/bin/systemctl restart kiosk.service" {
		t.Fatalf("unexpected restart command: %#v", runner.calls)
	}
}

func TestSystemdControllerRestartFailureAndTimeout(t *testing.T) {
	runner := &fakeRunner{showOutput: activeService(), restartErr: errors.New("sudo denied")}
	controller := NewSystemdServiceController("chromium")
	controller.runner, controller.inspectPID = runner, kioskProcess
	if err := controller.Restart(context.Background()); err == nil {
		t.Fatal("expected restart failure")
	}
	runner.restartErr = nil
	runner.showOutput = "LoadState=loaded\nActiveState=activating\nSubState=start\nMainPID=0\n"
	controller.timeout, controller.poll = 5*time.Millisecond, time.Millisecond
	if err := controller.Restart(context.Background()); err == nil {
		t.Fatal("expected restart timeout")
	}
}
