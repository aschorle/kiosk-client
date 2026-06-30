#!/bin/sh
#
# Start the Cage runtime for kiosk-client.
#
# Purpose:
#   Starts Cage and runs scripts/start-browser.sh inside it. This script owns
#   the appliance runtime path.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
CAGE_BIN=${CAGE_BIN:-cage}
BROWSER_SCRIPT=$PROJECT_DIR/scripts/start-browser.sh

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

	if [ ! -x "$BROWSER_SCRIPT" ]; then
		log_error "Browser-Startskript ist nicht ausfuehrbar: $BROWSER_SCRIPT"
		return 1
	fi

	if ! cd "$PROJECT_DIR"; then
		log_error "Working Directory konnte nicht gesetzt werden: $PROJECT_DIR"
		return 1
	fi

	log_info "Working Directory: $PROJECT_DIR"
	log_info "Starte Cage: $cage_path"
	log_info "Starte Browser in Cage: $BROWSER_SCRIPT"

	exec "$cage_path" -- "$BROWSER_SCRIPT"
}

main() {
	start_cage
}

main "$@"
