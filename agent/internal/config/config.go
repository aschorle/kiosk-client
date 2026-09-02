package config

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	defaultURL                 = "http://localhost"
	defaultBrowser             = "chromium"
	defaultClientType          = "appliance"
	defaultBrowserController   = "supervisor"
	defaultWatchdogMode        = "restart"
	defaultHTTPAddr            = "127.0.0.1:8080"
	systemdClientType          = "systemd"
	systemdBrowserController   = "systemd-service"
	systemdBrowserService      = "kiosk.service"
	managementOnlyWatchdogMode = "observe"
)

// Config contains the kiosk-client runtime configuration.
type Config struct {
	URL               string `json:"url"`
	DeviceID          string `json:"device_id"`
	DeviceName        string `json:"device_name"`
	ServerURL         string `json:"server_url"`
	Browser           string `json:"browser"`
	AuthToken         string `json:"-"`
	ClientType        string `json:"client_type"`
	BrowserController string `json:"browser_controller"`
	BrowserService    string `json:"browser_service,omitempty"`
	WatchdogMode      string `json:"watchdog_mode"`
	EnableReboot      bool   `json:"enable_reboot"`
	HTTPAddr          string `json:"http_addr"`
	ConfigWritable    bool   `json:"config_writable"`
}

var (
	mu            sync.RWMutex
	currentConfig Config
	currentError  error
	currentPath   string
	loaded        bool
)

// ValidationError describes invalid user supplied configuration.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// Load reads config/client.conf style KEY=value configuration.
func Load(path string) (Config, error) {
	mu.Lock()
	defer mu.Unlock()

	if loaded {
		return currentConfig, nil
	}

	currentPath = path
	cfg, err := read(path)
	if err != nil {
		currentConfig = defaultConfig()
		currentError = err
		loaded = true
		return currentConfig, nil
	}

	currentConfig = cfg
	currentError = nil
	loaded = true
	return currentConfig, nil
}

// Current returns the configuration that was loaded during agent startup.
func Current() (Config, error) {
	mu.RLock()
	defer mu.RUnlock()

	if !loaded {
		return Config{}, fmt.Errorf("configuration has not been loaded")
	}

	if currentError != nil {
		return Config{}, currentError
	}

	return currentConfig, nil
}

// AuthToken returns the configured API token without exposing it through JSON.
func AuthToken() string {
	mu.RLock()
	defer mu.RUnlock()

	return strings.TrimSpace(currentConfig.AuthToken)
}

// ManagementStatePath returns the separate, non-secret state file beside the
// loaded client.conf.
func ManagementStatePath() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if !loaded || currentPath == "" {
		return "", fmt.Errorf("configuration has not been loaded")
	}
	return filepath.Join(filepath.Dir(currentPath), "management-state.json"), nil
}

// Update validates and writes the runtime configuration to client.conf.
func Update(cfg Config) error {
	mu.Lock()
	defer mu.Unlock()

	if !loaded {
		return fmt.Errorf("configuration has not been loaded")
	}
	if !currentConfig.ConfigWritable {
		return fmt.Errorf("configuration is read-only")
	}
	// The local dashboard may update only user-facing appliance values. Keep
	// controller and privilege settings local to the deployed profile.
	cfg.ClientType = currentConfig.ClientType
	cfg.BrowserController = currentConfig.BrowserController
	cfg.BrowserService = currentConfig.BrowserService
	cfg.WatchdogMode = currentConfig.WatchdogMode
	cfg.EnableReboot = currentConfig.EnableReboot
	cfg.HTTPAddr = currentConfig.HTTPAddr
	cfg.ConfigWritable = currentConfig.ConfigWritable
	normalized, err := Validate(cfg)
	if err != nil {
		return err
	}

	if currentPath == "" {
		return fmt.Errorf("configuration path is empty")
	}

	normalized.AuthToken = currentConfig.AuthToken
	normalized.DeviceName = currentConfig.DeviceName
	normalized.ServerURL = currentConfig.ServerURL

	mode := os.FileMode(0644)
	if info, err := os.Stat(currentPath); err == nil {
		mode = info.Mode().Perm()
	}

	content := fmt.Sprintf(
		"URL=%s\nDEVICE_ID=%s\nDEVICE_NAME=%s\nSERVER_URL=%s\nBROWSER=%s\nAUTH_TOKEN=%s\nCLIENT_TYPE=%s\nBROWSER_CONTROLLER=%s\nBROWSER_SERVICE=%s\nBROWSER_WATCHDOG=%s\nENABLE_REBOOT=%t\nHTTP_ADDR=%s\nCONFIG_WRITABLE=%t\n",
		normalized.URL,
		normalized.DeviceID,
		normalized.DeviceName,
		normalized.ServerURL,
		normalized.Browser,
		normalized.AuthToken,
		normalized.ClientType,
		normalized.BrowserController,
		normalized.BrowserService,
		normalized.WatchdogMode,
		normalized.EnableReboot,
		normalized.HTTPAddr,
		normalized.ConfigWritable,
	)
	if err := os.WriteFile(filepath.Clean(currentPath), []byte(content), mode); err != nil {
		return fmt.Errorf("write %s: %w", currentPath, err)
	}

	currentConfig = normalized
	currentError = nil
	return nil
}

// Validate normalizes and validates user supplied configuration values.
func Validate(cfg Config) (Config, error) {
	normalized := Config{
		URL:               strings.TrimSpace(cfg.URL),
		DeviceID:          strings.TrimSpace(cfg.DeviceID),
		DeviceName:        strings.TrimSpace(cfg.DeviceName),
		ServerURL:         strings.TrimRight(strings.TrimSpace(cfg.ServerURL), "/"),
		Browser:           strings.ToLower(strings.TrimSpace(cfg.Browser)),
		ClientType:        strings.ToLower(strings.TrimSpace(cfg.ClientType)),
		BrowserController: strings.ToLower(strings.TrimSpace(cfg.BrowserController)),
		BrowserService:    strings.TrimSpace(cfg.BrowserService),
		WatchdogMode:      strings.ToLower(strings.TrimSpace(cfg.WatchdogMode)),
		EnableReboot:      cfg.EnableReboot,
		HTTPAddr:          strings.TrimSpace(cfg.HTTPAddr),
		ConfigWritable:    cfg.ConfigWritable,
	}

	if normalized.URL == "" {
		return Config{}, ValidationError{Message: "url must not be empty"}
	}

	if normalized.Browser != defaultBrowser {
		return Config{}, ValidationError{Message: "browser must be chromium"}
	}
	if normalized.ClientType == "" {
		normalized.ClientType = defaultClientType
	}
	if normalized.ClientType != defaultClientType && normalized.ClientType != systemdClientType {
		return Config{}, ValidationError{Message: "client_type must be appliance or systemd"}
	}
	if normalized.ClientType == systemdClientType {
		if normalized.BrowserController == "" {
			normalized.BrowserController = systemdBrowserController
		}
		if normalized.BrowserController != systemdBrowserController {
			return Config{}, ValidationError{Message: "systemd client_type requires systemd-service browser controller"}
		}
		if normalized.BrowserService == "" {
			normalized.BrowserService = systemdBrowserService
		}
		if normalized.BrowserService != systemdBrowserService {
			return Config{}, ValidationError{Message: "browser_service must be kiosk.service"}
		}
		if normalized.WatchdogMode == "" {
			normalized.WatchdogMode = managementOnlyWatchdogMode
		}
		if normalized.WatchdogMode != managementOnlyWatchdogMode {
			return Config{}, ValidationError{Message: "systemd client_type requires observe watchdog mode"}
		}
		if normalized.EnableReboot {
			return Config{}, ValidationError{Message: "systemd client_type does not support reboot"}
		}
		if normalized.HTTPAddr == "" {
			normalized.HTTPAddr = "127.0.0.1:18080"
		}
		if normalized.ConfigWritable {
			return Config{}, ValidationError{Message: "systemd client_type requires read-only configuration"}
		}
	} else {
		if normalized.BrowserController == "" {
			normalized.BrowserController = defaultBrowserController
		}
		if normalized.BrowserController != defaultBrowserController {
			return Config{}, ValidationError{Message: "appliance client_type requires supervisor browser controller"}
		}
		if normalized.WatchdogMode == "" {
			normalized.WatchdogMode = defaultWatchdogMode
		}
		if normalized.WatchdogMode != defaultWatchdogMode && normalized.WatchdogMode != "off" {
			return Config{}, ValidationError{Message: "watchdog_mode must be restart or off"}
		}
		if normalized.HTTPAddr == "" {
			normalized.HTTPAddr = defaultHTTPAddr
		}
	}
	if err := validateLoopbackAddr(normalized.HTTPAddr); err != nil {
		return Config{}, ValidationError{Message: err.Error()}
	}

	return normalized, nil
}

func read(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("invalid config line %d: missing '='", lineNumber)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return Config{}, fmt.Errorf("invalid config line %d: empty key", lineNumber)
		}

		values[key] = strings.TrimSpace(value)
	}

	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	configWritable, err := parseBool(values["CONFIG_WRITABLE"], true)
	if err != nil {
		return Config{}, err
	}
	enableReboot, err := parseBool(values["ENABLE_REBOOT"], strings.ToLower(strings.TrimSpace(values["CLIENT_TYPE"])) != systemdClientType)
	if err != nil {
		return Config{}, err
	}
	cfg, err := Validate(Config{
		URL:               valueOrDefaultValue(values["URL"], defaultURL),
		DeviceID:          values["DEVICE_ID"],
		DeviceName:        values["DEVICE_NAME"],
		ServerURL:         strings.TrimRight(values["SERVER_URL"], "/"),
		Browser:           valueOrDefaultValue(values["BROWSER"], defaultBrowser),
		AuthToken:         values["AUTH_TOKEN"],
		ClientType:        values["CLIENT_TYPE"],
		BrowserController: values["BROWSER_CONTROLLER"],
		BrowserService:    values["BROWSER_SERVICE"],
		WatchdogMode:      values["BROWSER_WATCHDOG"],
		EnableReboot:      enableReboot,
		HTTPAddr:          values["HTTP_ADDR"],
		ConfigWritable:    configWritable,
	})
	if err != nil {
		return Config{}, err
	}
	cfg.AuthToken = values["AUTH_TOKEN"]
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		URL:               defaultURL,
		Browser:           defaultBrowser,
		AuthToken:         "",
		ClientType:        defaultClientType,
		BrowserController: defaultBrowserController,
		WatchdogMode:      defaultWatchdogMode,
		EnableReboot:      true,
		HTTPAddr:          defaultHTTPAddr,
		ConfigWritable:    true,
	}
}

func valueOrDefaultValue(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

func parseBool(value string, defaultValue bool) (bool, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("invalid boolean value %q", value)
	}
	return parsed, nil
}

func validateLoopbackAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil || host != "127.0.0.1" || port == "" {
		return fmt.Errorf("http_addr must bind to 127.0.0.1:<port>")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("http_addr must contain a valid port")
	}
	return nil
}
