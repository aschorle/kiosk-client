#!/bin/sh
#
# Radxa Rock 4C+ Appliance Edition installer workflow.
#
# Purpose:
#   Performs the Radxa Rock 4C+ Appliance installation workflow for
#   kiosk-client. This file only verifies the target board and delegates to the
#   minimal Appliance installer. It does not run Desktop Edition modules.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

on_error() {
	log_error "Radxa-Appliance-Installation abgebrochen."
}

trap on_error INT TERM HUP

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

ensure_radxa_board() {
	if ! board="$(detect_board)"; then
		return 1
	fi

	if [ "$board" = "radxa-rock-4c-plus" ]; then
		log_success "Radxa Rock 4C+ erkannt."
		return 0
	fi

	log_error "Dieses Skript ist nur fuer Radxa Rock 4C+ vorgesehen. Erkannt: $board."
	return 1
}

print_welcome() {
	log_info "Starte Appliance-Installation fuer kiosk-client auf Radxa Rock 4C+."
	log_info "Es werden keine Desktop-Komponenten installiert oder konfiguriert."
}

run_appliance_installer() {
	sh "$SCRIPT_DIR/appliance.sh"
}

install_radxa() {
	run_step "Root-Rechte pruefen" require_root
	run_step "Board erkennen" ensure_radxa_board
	print_welcome
	run_step "Appliance-Installer ausfuehren" run_appliance_installer
	log_success "Radxa-Appliance-Installation erfolgreich abgeschlossen."
}

main() {
	install_radxa
}

if [ "${0##*/}" = "install-radxa.sh" ]; then
	main "$@"
fi
