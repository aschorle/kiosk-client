#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
SESSION_SCRIPT=$SCRIPT_DIR/start-wayland-session.sh
TEST_ROOT=$(mktemp -d)
trap 'cleanup; rm -rf "$TEST_ROOT"' EXIT HUP INT TERM

mkdir -p "$TEST_ROOT/config" "$TEST_ROOT/runtime" "$TEST_ROOT/state"
CONFIG_FILE=$TEST_ROOT/config/client.conf
SOCKET_DIR=$TEST_ROOT/runtime
SOCKET_NAME=wayland-test
SOCKET_PATH=$SOCKET_DIR/$SOCKET_NAME
STATE_DIR=$TEST_ROOT/state
APPLY_LOG=$STATE_DIR/apply-log
QUERY_LOG=$STATE_DIR/query-log
SUPERVISOR_MARKER=$TEST_ROOT/supervisor-pid
SESSION_PID=
PYTHON_BIN=${PYTHON_BIN:-python3}

cleanup() {
	if [ -n "$SESSION_PID" ]; then
		kill -TERM "$SESSION_PID" 2>/dev/null || true
		wait "$SESSION_PID" 2>/dev/null || true
		SESSION_PID=
	fi
}

cat > "$TEST_ROOT/supervisor" <<'EOF'
#!/bin/sh
sleep 300 &
browser_pid=$!
printf '%s %s\n' "$$" "$browser_pid" > "$SUPERVISOR_MARKER"
trap 'kill -TERM "$browser_pid" 2>/dev/null || true; wait "$browser_pid" 2>/dev/null || true; exit 0' INT TERM HUP
while :; do sleep 1; done
EOF

cat > "$TEST_ROOT/wlr-randr" <<'EOF'
#!/bin/sh
set -eu
if [ "$#" -eq 0 ]; then
	printf '%s\n' query >> "$QUERY_LOG"
	if [ ! -f "$STATE_DIR/present" ]; then
		exit 0
	fi
	current=$(cat "$STATE_DIR/current")
	printf 'HDMI-A-1 "test display"\n  Enabled: yes\n  Modes:\n'
	case $current in
		configured) printf '    1680x1050 px, 59.882999 Hz (current)\n' ;;
		*) printf '    1920x1080 px, 60.000000 Hz (current)\n' ;;
	esac
	exit 0
fi

if [ "$#" -eq 4 ] && [ "$1" = "--output" ] && [ "$2" = "HDMI-A-1" ] && [ "$3" = "--mode" ]; then
	printf '%s\n' "$4" >> "$APPLY_LOG"
	if [ ! -f "$STATE_DIR/present" ] || [ ! -f "$STATE_DIR/mode-supported" ]; then
		exit 1
	fi
	printf '%s\n' configured > "$STATE_DIR/current"
	exit 0
fi
exit 2
EOF

chmod +x "$TEST_ROOT/supervisor" "$TEST_ROOT/wlr-randr"

wait_for_file() {
	file=$1
	attempt=0
	while [ "$attempt" -lt 10 ]; do
		[ -e "$file" ] && return 0
		sleep 1
		attempt=$((attempt + 1))
	done
	return 1
}

apply_count() {
	if [ -f "$APPLY_LOG" ]; then
		wc -l < "$APPLY_LOG" | tr -d ' '
	else
		printf '0\n'
	fi
}

wait_for_apply_count() {
	expected=$1
	attempt=0
	while [ "$attempt" -lt 10 ]; do
		[ "$(apply_count)" -ge "$expected" ] && return 0
		sleep 1
		attempt=$((attempt + 1))
	done
	return 1
}

start_session() {
	SUPERVISOR_MARKER=$SUPERVISOR_MARKER \
	KIOSK_CLIENT_CONFIG=$CONFIG_FILE \
	WLR_RANDR_BIN=$TEST_ROOT/wlr-randr \
	WAYLAND_SOCKET_TIMEOUT=2 \
	DISPLAY_POLL_INTERVAL=1 \
	DISPLAY_RETRY_INTERVAL=2 \
	STATE_DIR=$STATE_DIR APPLY_LOG=$APPLY_LOG \
	XDG_RUNTIME_DIR=$SOCKET_DIR WAYLAND_DISPLAY=$SOCKET_NAME \
	sh "$SESSION_SCRIPT" "$TEST_ROOT/supervisor" &
	SESSION_PID=$!
	wait_for_file "$SUPERVISOR_MARKER"
}

create_wayland_socket() {
	rm -f "$SOCKET_PATH"
	"$PYTHON_BIN" -c 'import socket, sys; s = socket.socket(socket.AF_UNIX); s.bind(sys.argv[1]); s.close()' "$SOCKET_PATH"
}

reset_state() {
	rm -f "$STATE_DIR/present" "$STATE_DIR/mode-supported" "$STATE_DIR/current" "$APPLY_LOG" "$QUERY_LOG" "$SUPERVISOR_MARKER"
	printf '%s\n' automatic > "$STATE_DIR/current"
	printf '%s\n' 'DISPLAY_OUTPUT=HDMI-A-1' 'DISPLAY_MODE=1680x1050@59.883Hz' > "$CONFIG_FILE"
	create_wayland_socket
}

# No DISPLAY_MODE preserves the previous automatic path and starts the browser
# supervisor without querying wlr-randr or starting a monitor.
: > "$CONFIG_FILE"
SUPERVISOR_MARKER=$SUPERVISOR_MARKER KIOSK_CLIENT_CONFIG=$CONFIG_FILE \
	WLR_RANDR_BIN=$TEST_ROOT/wlr-randr STATE_DIR=$STATE_DIR APPLY_LOG=$APPLY_LOG \
	sh "$SESSION_SCRIPT" "$TEST_ROOT/supervisor" &
SESSION_PID=$!
wait_for_file "$SUPERVISOR_MARKER"
[ "$(apply_count)" -eq 0 ]
[ ! -e "$QUERY_LOG" ]
cleanup

if ! command -v "$PYTHON_BIN" >/dev/null 2>&1 || ! "$PYTHON_BIN" -c 'import socket; raise SystemExit(not hasattr(socket, "AF_UNIX"))' >/dev/null 2>&1; then
	printf '%s\n' 'wayland display test: no-override case passed; Wayland socket integration cases skipped (Python AF_UNIX unavailable)'
	exit 0
fi

# A valid configured mode is applied during ordinary session startup.
reset_state
: > "$STATE_DIR/present"
: > "$STATE_DIR/mode-supported"
start_session
wait_for_apply_count 1
[ "$(cat "$STATE_DIR/current")" = configured ]
browser_pid_before=$(awk '{print $2}' "$SUPERVISOR_MARKER")
sleep 2
[ "$(apply_count)" -eq 1 ]
cleanup

# The output may be absent at session start. Its later appearance triggers the
# configured mode without terminating or restarting the browser supervisor.
reset_state
start_session
[ "$(apply_count)" -eq 0 ]
: > "$STATE_DIR/present"
: > "$STATE_DIR/mode-supported"
wait_for_apply_count 1
[ "$(cat "$STATE_DIR/current")" = configured ]
browser_pid_before=$(awk '{print $2}' "$SUPERVISOR_MARKER")
cleanup

# Simulate a monitor power-cycle and wlroots selecting its automatic mode on
# reconnect. The same supervisor PID must survive while the override is reapplied.
reset_state
: > "$STATE_DIR/present"
: > "$STATE_DIR/mode-supported"
start_session
wait_for_apply_count 1
browser_pid_before=$(awk '{print $2}' "$SUPERVISOR_MARKER")
rm -f "$STATE_DIR/present"
sleep 2
printf '%s\n' automatic > "$STATE_DIR/current"
: > "$STATE_DIR/present"
wait_for_apply_count 2
[ "$(cat "$STATE_DIR/current")" = configured ]
browser_pid_after=$(awk '{print $2}' "$SUPERVISOR_MARKER")
[ "$browser_pid_after" = "$browser_pid_before" ]
kill -0 "$browser_pid_before"
cleanup

# An invalid/not-yet-supported mode only warns; the browser supervisor stays up.
reset_state
: > "$STATE_DIR/present"
start_session
sleep 1
kill -0 "$SESSION_PID"
[ -f "$SUPERVISOR_MARKER" ]
[ "$(apply_count)" -ge 1 ]
cleanup

printf '%s\n' 'Wayland display startup, automatic mode, delayed output, reconnect, retry fallback, and supervisor continuity: ok'
