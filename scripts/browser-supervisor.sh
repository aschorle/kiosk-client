#!/bin/sh
#
# Supervise Chromium inside the running Cage session.
#
# Purpose:
#   Keep Cage alive while Chromium is reloaded or restarted. The supervisor is
#   the long-lived Cage child and starts scripts/start-browser.sh as its child.

set -u

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
BROWSER_SCRIPT=$PROJECT_DIR/scripts/start-browser.sh
RUNTIME_DIR=${XDG_RUNTIME_DIR:-/tmp}/kiosk-client
PID_FILE=$RUNTIME_DIR/browser-supervisor.pid

browser_pid=""
shutdown_requested=0
restart_requested=0

log_info() {
	printf '[INFO] %s\n' "$*"
}

log_error() {
	printf '[ERROR] %s\n' "$*" >&2
}

write_pid_file() {
	if ! mkdir -p "$RUNTIME_DIR"; then
		log_error "Runtime-Verzeichnis konnte nicht erstellt werden: $RUNTIME_DIR"
		return 1
	fi

	if ! printf '%s\n' "$$" > "$PID_FILE"; then
		log_error "Supervisor-PID konnte nicht geschrieben werden: $PID_FILE"
		return 1
	fi
}

remove_pid_file() {
	if [ -f "$PID_FILE" ]; then
		rm -f "$PID_FILE"
	fi
}

start_browser() {
	if [ ! -x "$BROWSER_SCRIPT" ]; then
		log_error "Browser-Startskript ist nicht ausfuehrbar: $BROWSER_SCRIPT"
		return 1
	fi

	log_info "Starte Browser: $BROWSER_SCRIPT"
	"$BROWSER_SCRIPT" &
	browser_pid=$!
	log_info "Browser-PID: $browser_pid"
}

stop_browser() {
	if [ "${browser_pid:-}" = "" ]; then
		return 0
	fi

	if ! kill -0 "$browser_pid" 2>/dev/null; then
		return 0
	fi

	log_info "Beende Browser-PID: $browser_pid"
	kill -TERM "$browser_pid" 2>/dev/null || true

	attempt=1
	while [ "$attempt" -le 50 ]; do
		if ! kill -0 "$browser_pid" 2>/dev/null; then
			return 0
		fi

		sleep 0.1
		attempt=$((attempt + 1))
	done

	log_error "Browser reagiert nicht auf SIGTERM, sende SIGKILL: $browser_pid"
	kill -KILL "$browser_pid" 2>/dev/null || true
}

request_reload() {
	log_info "Browser Reload angefordert."
	restart_requested=1
	stop_browser
}

request_restart() {
	log_info "Browser Neustart angefordert."
	restart_requested=1
	stop_browser
}

request_shutdown() {
	log_info "Browser-Supervisor wird beendet."
	shutdown_requested=1
	stop_browser
}

cleanup() {
	trap - INT TERM USR1 USR2 EXIT
	request_shutdown
	remove_pid_file
}

run_supervisor() {
	if ! write_pid_file; then
		return 1
	fi

	trap request_reload USR1
	trap request_restart USR2
	trap request_shutdown INT TERM
	trap cleanup EXIT

	if ! cd "$PROJECT_DIR"; then
		log_error "Working Directory konnte nicht gesetzt werden: $PROJECT_DIR"
		return 1
	fi

	while [ "$shutdown_requested" -eq 0 ]; do
		restart_requested=0

		if ! start_browser; then
			return 1
		fi

		wait "$browser_pid"
		browser_status=$?
		browser_pid=""

		if [ "$shutdown_requested" -ne 0 ]; then
			break
		fi

		if [ "$restart_requested" -ne 0 ]; then
			log_info "Browser wird neu gestartet."
			continue
		fi

		log_error "Browser wurde unerwartet beendet: exit status $browser_status"
		sleep 1
	done
}

main() {
	run_supervisor
}

main "$@"
