#!/bin/sh
#
# Apply an optional output mode inside Cage's Wayland client session and keep
# it applied across output hotplug without restarting the browser supervisor.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
CONFIG_FILE=${KIOSK_CLIENT_CONFIG:-"$PROJECT_DIR/config/client.conf"}
SUPERVISOR_SCRIPT=${1:-"$SCRIPT_DIR/browser-supervisor.sh"}
WLR_RANDR_BIN=${WLR_RANDR_BIN:-wlr-randr}
WAYLAND_SOCKET_TIMEOUT=${WAYLAND_SOCKET_TIMEOUT:-10}
DISPLAY_POLL_INTERVAL=${DISPLAY_POLL_INTERVAL:-5}
DISPLAY_RETRY_INTERVAL=${DISPLAY_RETRY_INTERVAL:-15}

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

valid_interval() {
	case $1 in
		''|*[!0-9]*|0) return 1 ;;
		*) return 0 ;;
	esac
}

output_state() {
	# wlr-randr reports each output as an unindented heading followed by
	# properties. A disconnected output may be absent or have Enabled: no.
	output_info=$1
	printf '%s\n' "$output_info" | awk -v wanted="$display_output" '
		$1 == wanted { in_output = 1; next }
		in_output && /^[^[:space:]]/ { exit }
		in_output && $1 == "Enabled:" && $2 == "yes" { enabled = 1 }
		END { exit !enabled }
	'
}

current_mode() {
	output_info=$1
	printf '%s\n' "$output_info" | awk -v wanted="$display_output" '
		$1 == wanted { in_output = 1; next }
		in_output && /^[^[:space:]]/ { exit }
		in_output && /\(current\)/ {
			line = $0
			sub(/^[[:space:]]+/, "", line)
			split(line, fields, ", ")
			resolution = fields[1]
			sub(/[[:space:]]*px$/, "", resolution)
			refresh = fields[2]
			sub(/[[:space:]]*Hz.*/, "", refresh)
			if (resolution != "" && refresh != "") {
				printf "%s@%.3fHz\n", resolution, refresh + 0
			}
			exit
		}
	'
}

current_mode_matches() {
	observed_mode=$1
	[ "$observed_mode" = "$requested_mode" ]
}

query_output() {
	if ! output_info=$("$WLR_RANDR_BIN" 2>/dev/null); then
		return 1
	fi
	output_state "$output_info" || return 1
	printf '%s\n' "$output_info"
}

apply_display_mode() {
	if "$WLR_RANDR_BIN" --output "$display_output" --mode "$display_mode" >/dev/null 2>&1; then
		log_info "Wayland-Ausgabe $display_output auf $display_mode gesetzt."
		return 0
	fi
	return 1
}

monitor_display_mode() {
	output_present=$1
	mode_applied=$2
	failure_logged=$3
	retry_ticks=$(( (DISPLAY_RETRY_INTERVAL + DISPLAY_POLL_INTERVAL - 1) / DISPLAY_POLL_INTERVAL ))
	retry_countdown=$retry_ticks

	while :; do
		sleep "$DISPLAY_POLL_INTERVAL"
		if ! output_info=$(query_output); then
			if [ "$output_present" -eq 1 ]; then
				log_info "Wayland-Ausgabe $display_output ist nicht verfügbar; warte auf Hotplug."
			fi
			output_present=0
			mode_applied=0
			failure_logged=0
			retry_countdown=0
			continue
		fi

		if [ "$output_present" -eq 0 ]; then
			log_info "Wayland-Ausgabe $display_output ist wieder verfügbar."
			output_present=1
			mode_applied=0
			failure_logged=0
			retry_countdown=0
		fi

		observed_mode=$(current_mode "$output_info")
		if current_mode_matches "$observed_mode"; then
			if [ "$mode_applied" -eq 0 ]; then
				log_info "Konfigurierter Wayland-Modus $display_mode ist aktiv."
			fi
			mode_applied=1
			failure_logged=0
			retry_countdown=$retry_ticks
			continue
		fi

		if [ "$mode_applied" -eq 1 ]; then
			log_info "Wayland-Ausgabe $display_output nutzt nicht mehr den konfigurierten Modus; stelle ihn erneut ein."
			mode_applied=0
			retry_countdown=0
		fi

		if [ "$retry_countdown" -gt 0 ]; then
			retry_countdown=$((retry_countdown - 1))
			continue
		fi

		if apply_display_mode; then
			mode_applied=1
			failure_logged=0
			retry_countdown=$retry_ticks
		elif [ "$failure_logged" -eq 0 ]; then
			log_warn "Modus $display_mode für $display_output ist derzeit nicht verfügbar; der Kiosk läuft mit der automatischen Moduswahl weiter."
			failure_logged=1
			retry_countdown=$retry_ticks
		else
			retry_countdown=$retry_ticks
		fi
	done
}

stop_children() {
	trap - HUP INT TERM
	if [ -n "${monitor_pid:-}" ]; then
		kill -TERM "$monitor_pid" 2>/dev/null || true
	fi
	if [ -n "${supervisor_pid:-}" ]; then
		kill -TERM "$supervisor_pid" 2>/dev/null || true
	fi
	if [ -n "${monitor_pid:-}" ]; then
		wait "$monitor_pid" 2>/dev/null || true
	fi
	if [ -n "${supervisor_pid:-}" ]; then
		wait "$supervisor_pid" 2>/dev/null || true
	fi
}

run_supervisor_with_monitor() {
	monitor_display_mode "$output_present" "$mode_applied" "$initial_failure_logged" &
	monitor_pid=$!
	"$SUPERVISOR_SCRIPT" &
	supervisor_pid=$!
	trap 'stop_children; exit 0' HUP INT TERM

	if wait "$supervisor_pid"; then
		supervisor_status=0
	else
		supervisor_status=$?
	fi
	stop_children
	exit "$supervisor_status"
}

start_supervisor_without_monitor() {
	if [ ! -x "$SUPERVISOR_SCRIPT" ]; then
		log_warn "Browser-Supervisor ist nicht ausfuehrbar: $SUPERVISOR_SCRIPT"
		exit 1
	fi
	exec "$SUPERVISOR_SCRIPT"
}

main() {
	display_mode=$(read_config_value DISPLAY_MODE)
	display_output=$(read_config_value DISPLAY_OUTPUT)

	# Preserve the old automatic path exactly when no mode is configured.
	if [ -z "$display_mode" ]; then
		start_supervisor_without_monitor
	fi
	if [ -z "$display_output" ]; then
		log_warn "DISPLAY_MODE ist gesetzt, aber DISPLAY_OUTPUT fehlt; verwende automatische Moduswahl."
		start_supervisor_without_monitor
	fi
	if ! valid_interval "$DISPLAY_POLL_INTERVAL"; then
		DISPLAY_POLL_INTERVAL=5
	fi
	if ! valid_interval "$DISPLAY_RETRY_INTERVAL"; then
		DISPLAY_RETRY_INTERVAL=15
	fi
	mode_resolution=${display_mode%%@*}
	mode_refresh=${display_mode#*@}
	mode_refresh=${mode_refresh%Hz}
	requested_mode=$(awk -v resolution="$mode_resolution" -v refresh="$mode_refresh" 'BEGIN { printf "%s@%.3fHz", resolution, refresh + 0 }')

	if ! socket_path=$(wayland_socket_path); then
		log_warn "Wayland-Socket-Umgebung fehlt; verwende automatische Moduswahl."
		start_supervisor_without_monitor
	fi
	if ! wait_for_wayland_socket "$socket_path"; then
		log_warn "Wayland-Socket wurde nicht rechtzeitig verfuegbar ($socket_path); verwende automatische Moduswahl."
		start_supervisor_without_monitor
	fi
	if ! command -v "$WLR_RANDR_BIN" >/dev/null 2>&1; then
		log_warn "wlr-randr wurde nicht gefunden; verwende automatische Moduswahl."
		start_supervisor_without_monitor
	fi
	if [ ! -x "$SUPERVISOR_SCRIPT" ]; then
		log_warn "Browser-Supervisor ist nicht ausfuehrbar: $SUPERVISOR_SCRIPT"
		exit 1
	fi

	output_present=0
	mode_applied=0
	initial_failure_logged=0
	if output_info=$(query_output); then
		output_present=1
		if current_mode_matches "$(current_mode "$output_info")"; then
			mode_applied=1
		else
			if apply_display_mode; then
				mode_applied=1
			else
				log_warn "Modus $display_mode fuer $display_output konnte beim Start nicht gesetzt werden; der Kiosk läuft mit der automatischen Moduswahl weiter."
				initial_failure_logged=1
			fi
		fi
	fi

	run_supervisor_with_monitor
}

main "$@"
