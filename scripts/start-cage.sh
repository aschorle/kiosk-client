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
CURSOR_THEME_NAME=${CURSOR_THEME_NAME:-kiosk-hidden}
CURSOR_SIZE=${CURSOR_SIZE:-1}
CURSOR_THEME_ROOT=${XDG_RUNTIME_DIR:-/tmp}/kiosk-client-cursors

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

install_hidden_cursor_theme() {
	cursor_dir=$CURSOR_THEME_ROOT/$CURSOR_THEME_NAME/cursors
	cursor_file=$cursor_dir/left_ptr

	if ! mkdir -p "$cursor_dir"; then
		log_error "Cursor-Theme-Verzeichnis konnte nicht erstellt werden: $cursor_dir"
		return 1
	fi

	# Minimal valid Xcursor file: 1x1 ARGB pixel with alpha 0.
	if ! printf '\130\143\165\162\020\000\000\000\000\000\001\000\001\000\000\000\002\000\375\377\001\000\000\000\034\000\000\000\044\000\000\000\002\000\375\377\001\000\000\000\001\000\000\000\001\000\000\000\001\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000' > "$cursor_file"; then
		log_error "Transparenter Cursor konnte nicht geschrieben werden: $cursor_file"
		return 1
	fi

	for cursor_name in default arrow pointer hand1 hand2 text xterm crosshair move watch progress; do
		ln -sf left_ptr "$cursor_dir/$cursor_name"
	done

	export XCURSOR_PATH=$CURSOR_THEME_ROOT${XCURSOR_PATH:+:$XCURSOR_PATH}
	export XCURSOR_THEME=$CURSOR_THEME_NAME
	export XCURSOR_SIZE=$CURSOR_SIZE

	log_info "Transparentes Cursor-Theme: $CURSOR_THEME_NAME"
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

	install_hidden_cursor_theme

	log_info "Working Directory: $PROJECT_DIR"
	log_info "Starte Cage: $cage_path"
	log_info "Starte Browser in Cage: $BROWSER_SCRIPT"

	exec "$cage_path" -- "$BROWSER_SCRIPT"
}

main() {
	start_cage
}

main "$@"
