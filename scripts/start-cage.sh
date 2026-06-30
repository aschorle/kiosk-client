#!/bin/sh
#
# Start the Cage runtime for kiosk-client.
#
# Purpose:
#   Starts Cage and runs scripts/start-browser.sh inside it. This is the first
#   runtime wrapper for the future GNOME-free kiosk mode.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
CAGE_BIN=${CAGE_BIN:-cage}

log_info() {
	printf '[INFO] %s\n' "$*"
}

log_error() {
	printf '[ERROR] %s\n' "$*" >&2
}

find_cage() {
	if command -v "$CAGE_BIN" >/dev/null 2>&1; then
		command -v "$CAGE_BIN"
		return 0
	fi

	return 1
}

start_cage() {
	if ! cage_path=$(find_cage); then
		log_error "Cage wurde nicht gefunden."
		return 1
	fi

	if [ ! -x "$PROJECT_DIR/scripts/start-browser.sh" ]; then
		log_error "Browser-Startskript ist nicht ausführbar: $PROJECT_DIR/scripts/start-browser.sh"
		return 1
	fi

	log_info "Starte Cage: $cage_path"
	log_info "Starte Browser in Cage: $PROJECT_DIR/scripts/start-browser.sh"

	exec "$cage_path" "$PROJECT_DIR/scripts/start-browser.sh"
}

main() {
	start_cage
}

main "$@"
