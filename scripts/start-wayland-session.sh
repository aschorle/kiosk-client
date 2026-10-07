#!/bin/sh
#
# Apply an optional output mode inside Cage's Wayland client session, then
# start the normal browser supervisor regardless of display-configuration errors.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
CONFIG_FILE=${KIOSK_CLIENT_CONFIG:-"$PROJECT_DIR/config/client.conf"}
SUPERVISOR_SCRIPT=${1:-"$SCRIPT_DIR/browser-supervisor.sh"}
WLR_RANDR_BIN=${WLR_RANDR_BIN:-wlr-randr}
WAYLAND_SOCKET_TIMEOUT=${WAYLAND_SOCKET_TIMEOUT:-10}

log_info() {
	printf '[INFO] %s\n' "$*"
}

log_warn() {
	printf '[WARN] %s\n' "$*" >&2
}

read_config_value() {
	key=$1
	[ -r "$CONFIG_FILE" ] || return 0
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

wayland_socket_path() {
	if [ -z "${WAYLAND_DISPLAY:-}" ]; then
		return 1
	fi
	case $WAYLAND_DISPLAY in
		/*) printf '%s\n' "$WAYLAND_DISPLAY" ;;
		*)
			if [ -z "${XDG_RUNTIME_DIR:-}" ]; then
				return 1
			fi
			printf '%s/%s\n' "$XDG_RUNTIME_DIR" "$WAYLAND_DISPLAY"
			;;
	esac
}

wait_for_wayland_socket() {
	socket_path=$1
	remaining=$WAYLAND_SOCKET_TIMEOUT
	while [ "$remaining" -gt 0 ]; do
		if [ -S "$socket_path" ]; then
			return 0
		fi
		sleep 1
		remaining=$((remaining - 1))
	done
	[ -S "$socket_path" ]
}

start_supervisor() {
	if [ ! -x "$SUPERVISOR_SCRIPT" ]; then
		log_warn "Browser-Supervisor ist nicht ausfuehrbar: $SUPERVISOR_SCRIPT"
		exit 1
	fi
	exec "$SUPERVISOR_SCRIPT"
}

apply_optional_display_mode() {
	display_mode=$(read_config_value DISPLAY_MODE)
	display_output=$(read_config_value DISPLAY_OUTPUT)

	# No mode means preserve Cage's existing automatic output selection exactly.
	if [ -z "$display_mode" ]; then
		return 0
	fi
	if [ -z "$display_output" ]; then
		log_warn "DISPLAY_MODE ist gesetzt, aber DISPLAY_OUTPUT fehlt; verwende automatische Moduswahl."
		return 0
	fi

	if ! socket_path=$(wayland_socket_path); then
		log_warn "Wayland-Socket-Umgebung fehlt; verwende automatische Moduswahl."
		return 0
	fi
	if ! wait_for_wayland_socket "$socket_path"; then
		log_warn "Wayland-Socket wurde nicht rechtzeitig verfuegbar ($socket_path); verwende automatische Moduswahl."
		return 0
	fi

	if ! command -v "$WLR_RANDR_BIN" >/dev/null 2>&1; then
		log_warn "wlr-randr wurde nicht gefunden; verwende automatische Moduswahl."
		return 0
	fi

	if "$WLR_RANDR_BIN" --output "$display_output" --mode "$display_mode"; then
		log_info "Wayland-Ausgabe $display_output auf $display_mode gesetzt."
	else
		log_warn "Modus $display_mode fuer $display_output konnte nicht gesetzt werden; verwende automatische Moduswahl."
	fi
}

main() {
	apply_optional_display_mode
	start_supervisor
}

main "$@"
