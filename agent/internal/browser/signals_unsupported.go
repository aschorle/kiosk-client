//go:build !linux

package browser

import (
	"fmt"
	"os"
)

var (
	supervisorSignalReload  os.Signal = unsupportedSignal("reload")
	supervisorSignalRestart os.Signal = unsupportedSignal("restart")
)

type unsupportedSignal string

func (s unsupportedSignal) String() string { return string(s) }
func (unsupportedSignal) Signal()          {}

func signalProcess(_ int, _ os.Signal) error {
	return fmt.Errorf("browser supervisor signals are supported on linux only")
}
