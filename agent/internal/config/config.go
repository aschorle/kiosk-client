package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	defaultURL     = "http://localhost"
	defaultBrowser = "chromium"
)

// Config contains the kiosk-client runtime configuration.
type Config struct {
	URL      string `json:"url"`
	DeviceID string `json:"device_id"`
	Browser  string `json:"browser"`
}

var (
	currentConfig Config
	currentError  error
	loaded        bool
)

// Load reads config/client.conf style KEY=value configuration.
func Load(path string) (Config, error) {
	if loaded {
		return currentConfig, nil
	}

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
	if !loaded {
		return Config{}, fmt.Errorf("configuration has not been loaded")
	}

	if currentError != nil {
		return Config{}, currentError
	}

	return currentConfig, nil
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
		URL:      valueOrDefault(values["URL"], defaultURL),
		DeviceID: values["DEVICE_ID"],
		Browser:  valueOrDefault(values["BROWSER"], defaultBrowser),
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		URL:     defaultURL,
		Browser: defaultBrowser,
	}
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
