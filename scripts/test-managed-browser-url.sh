#!/bin/sh
# Regression test for managed URL precedence and safe fallback behavior.
set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
START_SCRIPT=$SCRIPT_DIR/start-browser.sh
TEST_ROOT=$(mktemp -d)
trap 'rm -rf "$TEST_ROOT"' EXIT HUP INT TERM

mkdir -p "$TEST_ROOT/bin" "$TEST_ROOT/config"
printf '%s\n' 'URL=http://local.example/kiosk/' 'BROWSER=chromium' > "$TEST_ROOT/config/client.conf"
printf '%s\n' '#!/bin/sh' 'printf "%s\\n" "$@"' > "$TEST_ROOT/bin/chromium"
chmod 0755 "$TEST_ROOT/bin/chromium"

run_url() {
	KIOSK_CLIENT_CONFIG="$TEST_ROOT/config/client.conf" \
	KIOSK_CLIENT_MANAGEMENT_STATE="$TEST_ROOT/config/management-state.json" \
	PATH="$TEST_ROOT/bin:$PATH" \
		"$START_SCRIPT" 2>/dev/null | tail -n 1
}

printf '%s\n' '{"effective_browser_url":"http://example.test/player/","desired_revision":"1","applied_revision":"1","status":"synced"}' > "$TEST_ROOT/config/management-state.json"
[ "$(run_url)" = "http://example.test/player/" ]

rm -f "$TEST_ROOT/config/management-state.json"
[ "$(run_url)" = "http://local.example/kiosk/" ]

printf '%s\n' '{broken json' > "$TEST_ROOT/config/management-state.json"
[ "$(run_url)" = "http://local.example/kiosk/" ]

printf '%s\n' '{"desired_revision":"1","status":"synced"}' > "$TEST_ROOT/config/management-state.json"
[ "$(run_url)" = "http://local.example/kiosk/" ]

printf '%s\n' 'managed browser URL precedence and fallback: ok'
