#!/bin/sh
#
# Shared installer helpers for kiosk-client.
#
# Purpose:
#   Provides common logging and preflight checks used by the installer.
#   The functions in this file only inspect the current system. They do not
#   install packages, write configuration, enable services, or change state.

set -eu

# ANSI color definitions for human-readable terminal output.
# Colors are disabled automatically when stdout is not a terminal or NO_COLOR is set.
if [ -t 1 ] && [ "${NO_COLOR:-}" = "" ]; then
	COLOR_RESET="$(printf '\033[0m')"
	COLOR_INFO="$(printf '\033[34m')"
	COLOR_WARN="$(printf '\033[33m')"
	COLOR_ERROR="$(printf '\033[31m')"
	COLOR_SUCCESS="$(printf '\033[32m')"
else
	COLOR_RESET=""
	COLOR_INFO=""
	COLOR_WARN=""
	COLOR_ERROR=""
	COLOR_SUCCESS=""
fi

# Minimum free space for preflight checks, in KiB.
# The value is intentionally conservative for the current framework phase.
MIN_FREE_SPACE_KB="${MIN_FREE_SPACE_KB:-1048576}"

log_info() {
	# Print an informational message.
	printf '%s[INFO]%s %s\n' "$COLOR_INFO" "$COLOR_RESET" "$*"
}

log_warn() {
	# Print a warning message.
	printf '%s[WARN]%s %s\n' "$COLOR_WARN" "$COLOR_RESET" "$*" >&2
}

log_error() {
	# Print an error message.
	printf '%s[ERROR]%s %s\n' "$COLOR_ERROR" "$COLOR_RESET" "$*" >&2
}

log_success() {
	# Print a success message.
	printf '%s[OK]%s %s\n' "$COLOR_SUCCESS" "$COLOR_RESET" "$*"
}

require_root() {
	# Verify that the script runs with root privileges.
	# Returns 0 for root and 1 for non-root users.
	if [ "$(id -u)" -eq 0 ]; then
		log_success "Root-Rechte vorhanden."
		return 0
	fi

	log_error "Root-Rechte erforderlich."
	return 1
}

check_debian() {
	# Verify that the operating system is Debian or a Debian derivative on a
	# supported stable base. Returns 0 for supported systems and 1 otherwise.
	if [ ! -r /etc/os-release ]; then
		log_error "/etc/os-release nicht lesbar."
		return 1
	fi

	# shellcheck disable=SC1091
	. /etc/os-release

	case "${VERSION_CODENAME:-}" in
		bookworm|trixie)
			;;
		*)
			log_error "Nicht unterstuetzter Debian-Codename: ${VERSION_CODENAME:-unbekannt}."
			return 1
			;;
	esac

	case " ${ID:-} ${ID_LIKE:-} " in
		*" debian "*|*" armbian "*)
			log_success "Debian-kompatibles System erkannt: ${ID:-unbekannt} ${VERSION_CODENAME:-unbekannt}."
			return 0
			;;
	esac

	log_error "Nicht unterstuetztes Betriebssystem: ID=${ID:-unbekannt} ID_LIKE=${ID_LIKE:-unbekannt}."
	return 1
}

check_architecture() {
	# Verify that the CPU architecture matches supported ARM targets.
	# Returns 0 for supported architectures and 1 otherwise.
	architecture="$(uname -m)"

	case "$architecture" in
		aarch64|arm64|armv7l)
			log_success "Unterstützte CPU-Architektur erkannt: $architecture."
			return 0
			;;
		*)
			log_error "Nicht unterstützte CPU-Architektur: $architecture."
			return 1
			;;
	esac
}

check_network() {
	# Verify basic network name resolution.
	# This check avoids changing network configuration and only performs lookup tests.
	if command -v getent >/dev/null 2>&1; then
		if getent hosts debian.org >/dev/null 2>&1; then
			log_success "Netzwerkverbindung verfügbar."
			return 0
		fi

		log_error "Netzwerkprüfung fehlgeschlagen: debian.org nicht auflösbar."
		return 1
	fi

	log_warn "getent nicht verfügbar; Netzwerkprüfung kann nicht ausgeführt werden."
	return 1
}

check_diskspace() {
	# Verify that the root filesystem has enough free space for later phases.
	# The default threshold is MIN_FREE_SPACE_KB and can be overridden by environment.
	available_kb="$(df -Pk / | awk 'NR == 2 { print $4 }')"

	case "${available_kb:-}" in
		''|*[!0-9]*)
			log_error "Freier Speicherplatz konnte nicht ermittelt werden."
			return 1
			;;
	esac

	if [ "${available_kb:-0}" -ge "$MIN_FREE_SPACE_KB" ]; then
		log_success "Ausreichend freier Speicherplatz vorhanden: ${available_kb} KiB."
		return 0
	fi

	log_error "Zu wenig freier Speicherplatz: ${available_kb:-0} KiB verfügbar, ${MIN_FREE_SPACE_KB} KiB erforderlich."
	return 1
}

detect_board() {
	# Detect the supported target board from device-tree or CPU information.
	# Prints one of: radxa-rock-4c-plus, raspberry-pi-4.
	# Returns 1 when the board cannot be identified as supported.
	board_model=""

	if [ -r /proc/device-tree/model ]; then
		board_model="$(tr -d '\000' </proc/device-tree/model)"
	elif [ -r /sys/firmware/devicetree/base/model ]; then
		board_model="$(tr -d '\000' </sys/firmware/devicetree/base/model)"
	elif [ -r /proc/cpuinfo ]; then
		board_model="$(awk -F ': ' '/Model|Hardware/ { print $2; exit }' /proc/cpuinfo)"
	fi

	case "$board_model" in
		*"Radxa ROCK 4C+"*|*"ROCK 4C+"*)
			printf '%s\n' "radxa-rock-4c-plus"
			return 0
			;;
		*"Raspberry Pi 4"*)
			printf '%s\n' "raspberry-pi-4"
			return 0
			;;
		*)
			log_error "Nicht unterstütztes oder unbekanntes Board: ${board_model:-unbekannt}."
			return 1
			;;
	esac
}
