#!/bin/sh
#
# Management-only installer for an existing Ubuntu 24.04 amd64 kiosk.
# It intentionally installs only kiosk-agent files. It does not manage the
# display stack, Chromium, kiosk.service, nginx, or the legacy slideshow.

set -eu

KIOSK_USER=${KIOSK_USER:-aschorle}
AGENT_BINARY=${AGENT_BINARY:-}
AGENT_SHA256=${AGENT_SHA256:-}
KIOSK_AGENT_CONFIG=${KIOSK_AGENT_CONFIG:-}
ENABLE_AGENT=${ENABLE_AGENT:-0}
AGENT_ROOT=/opt/kiosk-agent
CONFIG_ROOT=/etc/kiosk-agent
UNIT_SOURCE=$(CDPATH= cd "$(dirname "$0")/../systemd" && pwd)/kiosk-agent-management-only.service
SUDOERS_SOURCE=$(CDPATH= cd "$(dirname "$0")/../systemd/sudoers.d" && pwd)/kiosk-agent-mini-browser

fail() {
	printf '%s\n' "[ERROR] $*" >&2
	exit 1
}

require_root() {
	[ "$(id -u)" -eq 0 ] || fail "Root privileges are required."
}

check_platform() {
	[ -r /etc/os-release ] || fail "/etc/os-release is not readable."
	# shellcheck disable=SC1091
	. /etc/os-release
	[ "${ID:-}" = "ubuntu" ] && [ "${VERSION_ID:-}" = "24.04" ] || fail "This installer supports Ubuntu 24.04 only."
	[ "$(uname -m)" = "x86_64" ] || fail "This installer supports amd64 only."
}

check_inputs() {
	[ "$KIOSK_USER" = "aschorle" ] || fail "Management-only profile is fixed to KIOSK_USER=aschorle."
	[ -n "$AGENT_BINARY" ] && [ -f "$AGENT_BINARY" ] || fail "AGENT_BINARY must name a prebuilt agent binary."
	[ -n "$KIOSK_AGENT_CONFIG" ] && [ -f "$KIOSK_AGENT_CONFIG" ] || fail "KIOSK_AGENT_CONFIG must name a prepared client.conf."
	printf '%s' "$AGENT_SHA256" | grep -Eq '^[0-9a-fA-F]{64}$' || fail "AGENT_SHA256 must be a SHA-256 checksum."
	actual_sha256=$(sha256sum "$AGENT_BINARY" | awk '{print $1}')
	[ "$actual_sha256" = "$AGENT_SHA256" ] || fail "Agent binary checksum mismatch."
	id "$KIOSK_USER" >/dev/null 2>&1 || fail "KIOSK_USER does not exist."
	[ -r "$UNIT_SOURCE" ] && [ -r "$SUDOERS_SOURCE" ] || fail "Installer templates are missing."
	if command -v ss >/dev/null 2>&1 && ss -ltnH | awk '{print $4}' | grep -Eq '(^|[:])18080$'; then
		fail "127.0.0.1:18080 is already in use."
	fi
	[ ! -e "$AGENT_ROOT/kiosk-agent" ] || fail "Existing agent binary found; use the versioned update procedure."
	[ ! -e "$CONFIG_ROOT/client.conf" ] || fail "Existing client configuration found; it will not be overwritten."
}

install_files() {
	install -d -o root -g root -m 0755 "$AGENT_ROOT"
	install -d -o root -g "$KIOSK_USER" -m 0750 "$CONFIG_ROOT"
	install -o root -g root -m 0755 "$AGENT_BINARY" "$AGENT_ROOT/kiosk-agent"
	install -o root -g "$KIOSK_USER" -m 0640 "$KIOSK_AGENT_CONFIG" "$CONFIG_ROOT/client.conf"
	install -o root -g root -m 0644 "$UNIT_SOURCE" /etc/systemd/system/kiosk-agent.service

	temporary_sudoers=$(mktemp)
	trap 'rm -f "$temporary_sudoers"' EXIT HUP INT TERM
	sed "s/^aschorle /$KIOSK_USER /" "$SUDOERS_SOURCE" >"$temporary_sudoers"
	visudo -cf "$temporary_sudoers" >/dev/null || fail "Generated sudoers rule is invalid."
	install -o root -g root -m 0440 "$temporary_sudoers" /etc/sudoers.d/kiosk-agent-mini-browser
	rm -f "$temporary_sudoers"
	trap - EXIT HUP INT TERM

	systemctl daemon-reload
	if [ "$ENABLE_AGENT" = "1" ]; then
		systemctl enable --now kiosk-agent.service
	fi
}

main() {
	require_root
	check_platform
	check_inputs
	install_files
	printf '%s\n' "[OK] Management-only agent files installed."
	if [ "$ENABLE_AGENT" != "1" ]; then
		printf '%s\n' "[INFO] Agent remains disabled. Set ENABLE_AGENT=1 only after explicit approval."
	fi
}

main "$@"
