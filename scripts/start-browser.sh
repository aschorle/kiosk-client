#!/bin/sh
#
# Start Chromium for the kiosk-client runtime.
#
# Purpose:
#   Launches Chromium in kiosk mode with a minimal, predictable set of runtime
#   flags. The URL currently defaults to http://localhost. Reading the URL from
#   config/client.conf is prepared here, but the configuration file is not
#   required yet.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
CONFIG_FILE=${KIOSK_CLIENT_CONFIG:-"$PROJECT_DIR/config/client.conf"}
DEFAULT_URL="http://localhost"

log_info() {
	printf '[INFO] %s\n' "$*"
}

log_error() {
	printf '[ERROR] %s\n' "$*" >&2
}

find_chromium() {
	# Resolve the Chromium executable path.
	# Debian Bookworm uses "chromium"; "chromium-browser" is accepted as a
	# compatibility fallback for images with alternate package naming.
	if command -v chromium >/dev/null 2>&1; then
		command -v chromium
		return 0
	fi

	if command -v chromium-browser >/dev/null 2>&1; then
		command -v chromium-browser
		return 0
	fi

	return 1
}

read_config_url() {
	# Read URL=... from config/client.conf when the file exists.
	# This intentionally supports only simple KEY=value syntax for now.
	if [ ! -r "$CONFIG_FILE" ]; then
		printf '%s\n' "$DEFAULT_URL"
		return 0
	fi

	config_url=$(awk -F '=' '
		/^[[:space:]]*#/ { next }
		/^[[:space:]]*URL[[:space:]]*=/ {
			value = $0
			sub(/^[^=]*=/, "", value)
			gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
			print value
			exit
		}
	' "$CONFIG_FILE")

	if [ "$config_url" != "" ]; then
		printf '%s\n' "$config_url"
		return 0
	fi

	printf '%s\n' "$DEFAULT_URL"
}

validate_url() {
	# Validate that the runtime URL is not empty and uses HTTP or HTTPS.
	url=$1

	case "$url" in
		http://*|https://*)
			return 0
			;;
		"")
			log_error "Keine Browser-URL angegeben."
			return 1
			;;
		*)
			log_error "Ungültige Browser-URL: $url"
			return 1
			;;
	esac
}

start_browser() {
	# Resolve Chromium, validate the URL, and replace this process with Chromium.
	if ! chromium_path=$(find_chromium); then
		log_error "Chromium wurde nicht gefunden."
		return 1
	fi

	if [ ! -x "$chromium_path" ]; then
		log_error "Chromium ist nicht ausführbar: $chromium_path"
		return 1
	fi

	if ! kiosk_url=$(read_config_url); then
		log_error "Browser-URL konnte nicht ermittelt werden."
		return 1
	fi

	if ! validate_url "$kiosk_url"; then
		return 1
	fi

	log_info "Starte Chromium: $chromium_path"
	log_info "Kiosk-URL: $kiosk_url"

	exec "$chromium_path" \
		--kiosk \
		--incognito \
		--no-first-run \
		--disable-session-crashed-bubble \
		--disable-infobars \
		--disable-features=Translate \
		--disable-sync \
		--overscroll-history-navigation=0 \
		"$kiosk_url"
}

main() {
	start_browser
}

main "$@"
