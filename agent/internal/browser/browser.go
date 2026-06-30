package browser

// Runtime describes the configured browser executable.
//
// The agent does not control Chromium yet. This package exists as the future
// boundary for browser runtime integration while keeping configuration and
// status reporting independent.
type Runtime struct {
	Name string
}

// NewRuntime returns a browser runtime descriptor without starting a process.
func NewRuntime(name string) Runtime {
	return Runtime{Name: name}
}
