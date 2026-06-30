#!/bin/sh
#
# Radxa Rock 4C+ installer workflow.
#
# Purpose:
#   Performs the Radxa Rock 4C+ installation workflow for kiosk-client. After
#   the base installation succeeds, this script dispatches the existing module
#   scripts in a fixed order so the complete installation chain can be tested.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/packages.sh"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/verify.sh"

on_error() {
	# Print one consistent final error message for unexpected failures.
	log_error "Radxa-Grundinstallation abgebrochen."
}

trap on_error INT TERM HUP

run_step() {
	# Run one installer step with consistent logging and error propagation.
	# Arguments:
	#   $1: Human-readable step description.
	#   $2...: Command or function with optional arguments.
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
	# Detect the current board and ensure this hardware-specific installer is
	# only used on a Radxa Rock 4C+.
	if ! board="$(detect_board)"; then
		return 1
	fi

	if [ "$board" = "radxa-rock-4c-plus" ]; then
		log_success "Radxa Rock 4C+ erkannt."
		return 0
	fi

	log_error "Dieses Skript ist nur für Radxa Rock 4C+ vorgesehen. Erkannt: $board."
	return 1
}

print_welcome() {
	# Print a short greeting after all preflight checks have passed.
	log_info "Starte Grundinstallation für kiosk-client auf Radxa Rock 4C+."
	log_info "Installiert werden nur Systemupdates und gemeinsame Basispakete."
}

apt_update() {
	# Refresh apt package metadata. This step is safe to run repeatedly.
	apt update
}

apt_full_upgrade() {
	# Upgrade already installed packages. The noninteractive environment avoids
	# unnecessary prompts during unattended base setup.
	DEBIAN_FRONTEND=noninteractive
	export DEBIAN_FRONTEND

	apt full-upgrade -y
}

install_common_packages() {
	# Install the shared base package set from packages.sh.
	# The package list is converted into positional parameters so each package
	# name is passed to apt as an individual argument.
	DEBIAN_FRONTEND=noninteractive
	export DEBIAN_FRONTEND

	set --
	for package in $COMMON_PACKAGES; do
		set -- "$@" "$package"
	done

	if [ "$#" -eq 0 ]; then
		log_warn "COMMON_PACKAGES ist leer; keine Pakete werden installiert."
		return 0
	fi

	apt install -y --no-install-recommends "$@"
}

run_module_phase() {
	# Execute one installer module as a separate script.
	# Arguments:
	#   $1: Phase name displayed in the log output.
	#   $2: Module script file name below installer/.
	phase_name=$1
	module_file=$2
	module_path=$SCRIPT_DIR/$module_file

	log_info "=== $phase_name ==="

	if [ ! -r "$module_path" ]; then
		log_error "Modul nicht lesbar: $module_path."
		return 1
	fi

	sh "$module_path"
}

install_radxa() {
	# Execute the Radxa base installation in the required order.
	run_step "Root-Rechte prüfen" require_root
	run_step "Vorabprüfungen starten" run_preflight_checks
	run_step "Board erkennen" ensure_radxa_board
	print_welcome
	run_step "apt update ausführen" apt_update
	run_step "apt full-upgrade ausführen" apt_full_upgrade
	run_step "Gemeinsame Basispakete installieren" install_common_packages
	log_success "Radxa-Grundinstallation erfolgreich abgeschlossen."
	run_step "Browser-Modul ausführen" run_module_phase "Browser" "browser.sh"
	run_step "Wayland-Modul ausführen" run_module_phase "Wayland" "wayland.sh"
	run_step "Systemd-Modul ausführen" run_module_phase "Systemd" "systemd.sh"
	run_step "Network-Modul ausführen" run_module_phase "Network" "network.sh"
	run_step "Cleanup-Modul ausführen" run_module_phase "Cleanup" "cleanup.sh"
	log_success "Radxa-Installationskette erfolgreich abgeschlossen."
}

main() {
	# Entry point for direct execution.
	install_radxa
}

if [ "${0##*/}" = "install-radxa.sh" ]; then
	main "$@"
fi
