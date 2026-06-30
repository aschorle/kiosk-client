#!/bin/sh
#
# Raspberry Pi 4 Appliance Edition installer workflow.
#
# Purpose:
#   Performs the Raspberry Pi 4 Appliance installation workflow for
#   kiosk-client. This file only verifies the target board and delegates to the
#   minimal Appliance installer.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

run_step() {
	description=$1
	shift

	log_info "$description"

	if "$@"; then
		log_success "$description abgeschlossen."
		return 0
	fi

	log_error "$description fehlgeschlagen."
	return 1
}

ensure_rpi_board() {
	if ! board="$(detect_board)"; then
		return 1
	fi

	if [ "$board" = "raspberry-pi-4" ]; then
		log_success "Raspberry Pi 4 erkannt."
		return 0
	fi

	log_error "Dieses Skript ist nur fuer Raspberry Pi 4 vorgesehen. Erkannt: $board."
	return 1
}

install_rpi() {
	run_step "Root-Rechte pruefen" require_root
	run_step "Board erkennen" ensure_rpi_board
	run_step "Appliance-Installer ausfuehren" sh "$SCRIPT_DIR/appliance.sh"
	log_success "Raspberry-Pi-Appliance-Installation erfolgreich abgeschlossen."
}

verify_rpi() {
	ensure_rpi_board
}

if [ "${0##*/}" = "install-rpi.sh" ]; then
	install_rpi
fi
