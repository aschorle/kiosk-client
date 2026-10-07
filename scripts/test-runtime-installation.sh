#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
RUNTIME_SCRIPT=$SCRIPT_DIR/../installer/runtime.sh
TEST_ROOT=$(mktemp -d)
trap 'rm -rf "$TEST_ROOT"' EXIT HUP INT TERM

if ! command -v bash >/dev/null 2>&1; then
	printf '%s\n' 'runtime installation test: skipped (bash unavailable)'
	exit 0
fi

# Source runtime.sh using an alternate argv[0] from its own directory. This
# loads its functions without invoking main(), while testing the actual code.
bash -c '
set -eu
runtime_script=$1
test_root=$2
source_dir=$3
chown() { return 0; }
. "$runtime_script"

PROJECT_DIR=$test_root/project
user_home=$test_root/home/kiosk
CHROMIUM_POLICY_DIR=$test_root/chromium/policies/managed
CHROMIUM_POLICY_FILE=$CHROMIUM_POLICY_DIR/kiosk-client.json
mkdir -p "$PROJECT_DIR/systemd/user" "$PROJECT_DIR/config"

cp "$source_dir/../systemd/user/kiosk-agent.service" "$PROJECT_DIR/systemd/user/kiosk-agent.service"
cp "$source_dir/../systemd/user/kiosk-appliance.service" "$PROJECT_DIR/systemd/user/kiosk-appliance.service"
printf "%s\n" "URL=https://example.invalid/" "BROWSER=chromium" > "$PROJECT_DIR/config/client.conf"

install_service_file kiosk "$user_home" kiosk-agent.service
install_service_file kiosk "$user_home" kiosk-appliance.service

agent_unit=$user_home/.config/systemd/user/kiosk-agent.service
appliance_unit=$user_home/.config/systemd/user/kiosk-appliance.service
grep -Fx "WorkingDirectory=$PROJECT_DIR" "$agent_unit" >/dev/null
grep -Fx "ExecStart=$PROJECT_DIR/kiosk-agent" "$agent_unit" >/dev/null
grep -Fx "WorkingDirectory=$PROJECT_DIR" "$appliance_unit" >/dev/null
grep -Fx "ExecStart=/usr/bin/dbus-run-session -- $PROJECT_DIR/scripts/start-cage.sh" "$appliance_unit" >/dev/null
! grep -F "@PROJECT_DIR@" "$agent_unit" "$appliance_unit" >/dev/null

configure_runtime_config kiosk
[ "$(stat -c "%a" "$PROJECT_DIR/config/client.conf")" = "644" ]
[ -w "$PROJECT_DIR/config/client.conf" ]

mkdir -p "$CHROMIUM_POLICY_DIR"
printf "%s\n" "{\"UnrelatedPolicy\":true}" > "$CHROMIUM_POLICY_DIR/other.json"
install_chromium_policy
[ "$(cat "$CHROMIUM_POLICY_FILE")" = "{\"TranslateEnabled\":false}" ]
[ "$(stat -c "%a" "$CHROMIUM_POLICY_FILE")" = "644" ]
[ "$(cat "$CHROMIUM_POLICY_DIR/other.json")" = "{\"UnrelatedPolicy\":true}" ]
' "$SCRIPT_DIR/../installer/runtime-functions" "$RUNTIME_SCRIPT" "$TEST_ROOT" "$SCRIPT_DIR"

printf '%s\n' 'runtime service paths, configuration permissions, and Chromium policy: ok'
