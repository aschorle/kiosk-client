#!/bin/sh
#
# Start Chromium for the kiosk-client runtime.
#
# Purpose:
#   Launches Chromium in kiosk mode with a minimal, predictable set of runtime
#   flags. Runtime values are read from config/client.conf.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
CONFIG_FILE=${KIOSK_CLIENT_CONFIG:-"$PROJECT_DIR/config/client.conf"}
DEFAULT_URL="http://localhost"
LOCAL_ADMIN_URL="http://localhost:8080/"
WELCOME_URL="http://localhost:8080/welcome"
DEFAULT_BROWSER="chromium"

log_info() {
	printf '[INFO] %s\n' "$*"
}

log_error() {
	printf '[ERROR] %s\n' "$*" >&2
}

require_config_file() {
	# Ensure the runtime configuration exists before Chromium is started.
	if [ -r "$CONFIG_FILE" ]; then
		return 0
	fi

	log_error "Konfigurationsdatei fehlt oder ist nicht lesbar: $CONFIG_FILE"
	return 1
}

read_config_value() {
	# Read a single KEY=value entry from config/client.conf.
	# Comments and empty lines are ignored. Values are trimmed but otherwise
	# passed through unchanged.
	key=$1

	awk -F '=' -v key="$key" '
		/^[[:space:]]*#/ { next }
		/^[[:space:]]*$/ { next }
		{
			name = $1
			gsub(/^[[:space:]]+|[[:space:]]+$/, "", name)
			if (name == key) {
				value = $0
				sub(/^[^=]*=/, "", value)
				gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
				print value
				exit
			}
		}
	' "$CONFIG_FILE"
}

find_chromium() {
	# Resolve the Chromium executable path.
	# Debian Bookworm uses "chromium"; "chromium-browser" is accepted for
	# images with alternate package naming.
	browser_command=$(read_config_value "BROWSER")

	if [ "$browser_command" = "" ]; then
		browser_command=$DEFAULT_BROWSER
	fi

	if command -v "$browser_command" >/dev/null 2>&1; then
		command -v "$browser_command"
		return 0
	fi

	if [ "$browser_command" != "chromium-browser" ] && command -v chromium-browser >/dev/null 2>&1; then
		command -v chromium-browser
		return 0
	fi

	return 1
}

read_config_url() {
	# Read URL=... from config/client.conf and fall back to the local welcome
	# page until a real kiosk target is configured.
	config_url=$(read_config_value "URL")

	if is_configured_url "$config_url"; then
		printf '%s\n' "$config_url"
		return 0
	fi

	printf '%s\n' "$WELCOME_URL"
}

is_configured_url() {
	url=${1:-}

	case "$url" in
		http://*|https://*)
			;;
		*)
			return 1
			;;
	esac

	case "$url" in
		http://localhost|http://localhost/|https://localhost|https://localhost/|http://127.0.0.1|http://127.0.0.1/|https://127.0.0.1|https://127.0.0.1/)
			return 1
			;;
	esac

	return 0
}

wait_for_welcome_page() {
	url=$1

	if [ "$url" != "$WELCOME_URL" ]; then
		return 0
	fi

	attempt=1
	while [ "$attempt" -le 60 ]; do
		if command -v wget >/dev/null 2>&1; then
			if wget -q -O /dev/null "$LOCAL_ADMIN_URL" >/dev/null 2>&1; then
				return 0
			fi
		elif command -v curl >/dev/null 2>&1; then
			if curl -sS -o /dev/null "$LOCAL_ADMIN_URL" >/dev/null 2>&1; then
				return 0
			fi
		elif is_local_admin_port_listening; then
			return 0
		else
			log_info "Warte auf lokale Administration: $LOCAL_ADMIN_URL"
		fi

		sleep 1
		attempt=$((attempt + 1))
	done

	log_error "Lokale Administration ist nicht erreichbar: $LOCAL_ADMIN_URL"
	return 1
}

is_local_admin_port_listening() {
	for table in /proc/net/tcp /proc/net/tcp6; do
		if [ ! -r "$table" ]; then
			continue
		fi

		if awk 'NR > 1 {
			split($2, address, ":")
			if (address[2] == "1F90" && $4 == "0A") {
				found = 1
			}
		}
		END { exit !found }' "$table"; then
			return 0
		fi
	done

	return 1
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
	if ! require_config_file; then
		return 1
	fi

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

	if ! wait_for_welcome_page "$kiosk_url"; then
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
		--disable-crash-reporter \
		--disable-breakpad \
		--disable-background-networking \
		--disable-background-timer-throttling \
		--disable-renderer-backgrounding \
		--disable-sync \
		--overscroll-history-navigation=0 \
		"$kiosk_url"
}

main() {
	start_browser
}

main "$@"
