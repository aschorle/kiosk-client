//go:build linux

package browser

import (
	"os"
	"syscall"
)

var (
	supervisorSignalReload  os.Signal = syscall.SIGUSR1
	supervisorSignalRestart os.Signal = syscall.SIGUSR2
)

func signalProcess(pid int, signal os.Signal) error {
	return syscall.Kill(pid, signal.(syscall.Signal))
}
