#!/bin/sh
#
# kiosk-client installer dispatcher.
#
# Purpose:
#   This script is the single supported entry point for the Appliance Edition.
#   It performs only global checks and board selection, then delegates all
#   board-specific installation work to the matching Appliance installer.
#   Only Appliance Edition modules are run from this path.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)

# Load shared logging and system check helpers. Sourcing this file also prepares
# color variables used by the log_* functions.
# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

run_required_check() {
	# Run one mandatory dispatcher check with consistent logging.
	# Arguments:
	#   $1: Human-readable check description.
	#   $2...: Function or command to execute.
	description=$1
	shift

	log_info "$description"

	if "$@"; then
		return 0
	fi

	log_error "$description fehlgeschlagen."
	return 1
}

detect_supported_board() {
	# Detect the current hardware once and return the normalized board id.
	# detect_board prints the board id on stdout, so callers can capture it.
	detect_board
}

dispatch_board_installer() {
	# Delegate to the board-specific installer. The dispatcher intentionally uses
	# sh to avoid relying on executable file permissions in a freshly cloned tree.
	board=$1

	case "$board" in
		radxa-rock-4c-plus)
			log_info "Starte Radxa Rock 4C+ Installer."
			sh "$SCRIPT_DIR/install-radxa.sh"
			;;
		raspberry-pi-3|raspberry-pi-4)
			log_info "Starte Raspberry Pi Installer."
			sh "$SCRIPT_DIR/install-rpi.sh"
			;;
		*)
			log_error "Keine Installationsroutine für Board '$board' vorhanden."
			return 1
			;;
	esac
}

main() {
	# Global installer flow. Keep this order explicit so future changes remain
	# easy to audit: load helpers, initialize logging, check prerequisites,
	# detect the board, dispatch.
	log_info "kiosk-client Appliance Installation gestartet."

	run_required_check "Root-Rechte prüfen" require_root
	run_required_check "CPU-Architektur prüfen" check_architecture
	run_required_check "Debian-Version prüfen" check_debian

	if board="$(detect_supported_board)"; then
		log_success "Board erkannt: $board."
	else
		log_error "Das Board konnte nicht als unterstützte Zielplattform erkannt werden."
		return 1
	fi

	dispatch_board_installer "$board"
}

main "$@"
