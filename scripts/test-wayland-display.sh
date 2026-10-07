#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
SESSION_SCRIPT=$SCRIPT_DIR/start-wayland-session.sh
TEST_ROOT=$(mktemp -d)
trap 'rm -rf "$TEST_ROOT"' EXIT HUP INT TERM

mkdir -p "$TEST_ROOT/config" "$TEST_ROOT/runtime"
CONFIG_FILE=$TEST_ROOT/config/client.conf
SOCKET_DIR=$TEST_ROOT/runtime
SOCKET_NAME=wayland-test
SOCKET_PATH=$SOCKET_DIR/$SOCKET_NAME
RANDR_LOG=$TEST_ROOT/randr-args
SUPERVISOR_MARKER=$TEST_ROOT/supervisor-ran
PYTHON_BIN=${PYTHON_BIN:-python3}

cat > "$TEST_ROOT/supervisor" <<'EOF'
#!/bin/sh
printf '%s\n' ran > "$SUPERVISOR_MARKER"
EOF
cat > "$TEST_ROOT/wlr-randr-ok" <<'EOF'
#!/bin/sh
printf '%s\n' "$@" > "$RANDR_LOG"
EOF
cat > "$TEST_ROOT/wlr-randr-fail" <<'EOF'
#!/bin/sh
exit 1
EOF
chmod +x "$TEST_ROOT/supervisor" "$TEST_ROOT/wlr-randr-ok" "$TEST_ROOT/wlr-randr-fail"

create_wayland_socket() {
	"$PYTHON_BIN" -c 'import socket, sys; s = socket.socket(socket.AF_UNIX); s.bind(sys.argv[1]); s.close()' "$SOCKET_PATH"
}

# No DISPLAY_MODE must skip wlr-randr and directly continue to the supervisor.
: > "$CONFIG_FILE"
rm -f "$SUPERVISOR_MARKER" "$RANDR_LOG"
SUPERVISOR_MARKER=$SUPERVISOR_MARKER KIOSK_CLIENT_CONFIG=$CONFIG_FILE \
	WLR_RANDR_BIN=$TEST_ROOT/wlr-randr-ok \
	sh "$SESSION_SCRIPT" "$TEST_ROOT/supervisor"
[ -f "$SUPERVISOR_MARKER" ]
[ ! -e "$RANDR_LOG" ]

if ! command -v "$PYTHON_BIN" >/dev/null 2>&1 || ! "$PYTHON_BIN" -c 'import socket; raise SystemExit(not hasattr(socket, "AF_UNIX"))' >/dev/null 2>&1; then
	printf '%s\n' 'wayland display test: automatic mode passed; configured-mode socket cases skipped (Python AF_UNIX unavailable)'
	exit 0
fi

# A configured mode is applied with the expected output and mode arguments.
printf '%s\n' 'DISPLAY_OUTPUT=HDMI-A-1' 'DISPLAY_MODE=1680x1050@59.883Hz' > "$CONFIG_FILE"
create_wayland_socket
rm -f "$SUPERVISOR_MARKER" "$RANDR_LOG"
SUPERVISOR_MARKER=$SUPERVISOR_MARKER KIOSK_CLIENT_CONFIG=$CONFIG_FILE \
XDG_RUNTIME_DIR=$SOCKET_DIR WAYLAND_DISPLAY=$SOCKET_NAME \
WLR_RANDR_BIN=$TEST_ROOT/wlr-randr-ok RANDR_LOG=$RANDR_LOG \
sh "$SESSION_SCRIPT" "$TEST_ROOT/supervisor"
[ -f "$SUPERVISOR_MARKER" ]
printf '%s\n' '--output' 'HDMI-A-1' '--mode' '1680x1050@59.883Hz' > "$TEST_ROOT/expected-args"
cmp "$TEST_ROOT/expected-args" "$RANDR_LOG"

# A rejected/failed mode is non-fatal and the supervisor still starts.
rm -f "$SUPERVISOR_MARKER"
SUPERVISOR_MARKER=$SUPERVISOR_MARKER KIOSK_CLIENT_CONFIG=$CONFIG_FILE \
XDG_RUNTIME_DIR=$SOCKET_DIR WAYLAND_DISPLAY=$SOCKET_NAME \
WLR_RANDR_BIN=$TEST_ROOT/wlr-randr-fail \
sh "$SESSION_SCRIPT" "$TEST_ROOT/supervisor"
[ -f "$SUPERVISOR_MARKER" ]

printf '%s\n' 'Wayland display override: default, configured mode, and failure fallback: ok'
