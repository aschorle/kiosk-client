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
	URL      string
	DeviceID string
	Browser  string
}

// Load reads config/client.conf style KEY=value configuration.
func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	values := make(map[string]string)
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

		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
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

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
