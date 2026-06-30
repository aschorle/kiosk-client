package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	defaultURL     = "http://localhost"
	defaultBrowser = "chromium"
)

// Config contains the kiosk-client runtime configuration.
type Config struct {
	URL       string `json:"url"`
	DeviceID  string `json:"device_id"`
	Browser   string `json:"browser"`
	AuthToken string `json:"-"`
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

// Update validates and writes the runtime configuration to client.conf.
func Update(cfg Config) error {
	normalized, err := Validate(cfg)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	if !loaded {
		return fmt.Errorf("configuration has not been loaded")
	}

	if currentPath == "" {
		return fmt.Errorf("configuration path is empty")
	}

	normalized.AuthToken = currentConfig.AuthToken

	mode := os.FileMode(0644)
	if info, err := os.Stat(currentPath); err == nil {
		mode = info.Mode().Perm()
	}

	content := fmt.Sprintf(
		"URL=%s\nDEVICE_ID=%s\nBROWSER=%s\nAUTH_TOKEN=%s\n",
		normalized.URL,
		normalized.DeviceID,
		normalized.Browser,
		normalized.AuthToken,
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
		URL:      strings.TrimSpace(cfg.URL),
		DeviceID: strings.TrimSpace(cfg.DeviceID),
		Browser:  strings.ToLower(strings.TrimSpace(cfg.Browser)),
	}

	if normalized.URL == "" {
		return Config{}, ValidationError{Message: "url must not be empty"}
	}

	if normalized.Browser != defaultBrowser {
		return Config{}, ValidationError{Message: "browser must be chromium"}
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

	cfg := Config{
		URL:       valueOrDefault(values["URL"], defaultURL),
		DeviceID:  values["DEVICE_ID"],
		Browser:   valueOrDefault(values["BROWSER"], defaultBrowser),
		AuthToken: values["AUTH_TOKEN"],
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		URL:       defaultURL,
		Browser:   defaultBrowser,
		AuthToken: "",
	}
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
