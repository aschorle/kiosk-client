#!/bin/sh
#
# Installer verification module for kiosk-client.
#
# Purpose:
#   Runs read-only preflight checks for the kiosk-client installer framework.
#   This file does not install packages, write configuration, enable services,
#   or change the system.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

verify_debian_version() {
	# Verify that the host runs Debian 12 Bookworm.
	check_debian
}

verify_root_rights() {
	# Verify that the current user has root privileges.
	require_root
}

verify_network_connection() {
	# Verify basic network connectivity by checking DNS resolution.
	check_network
}

verify_free_diskspace() {
	# Verify available free space on the root filesystem.
	check_diskspace
}

verify_cpu_architecture() {
	# Verify that the CPU architecture is supported by target platforms.
	check_architecture
}

verify_supported_board() {
	# Verify that the current board is one of the supported targets.
	if board="$(detect_board)"; then
		log_success "Unterstütztes Board erkannt: $board."
		return 0
	fi

	return 1
}

run_preflight_checks() {
	# Run all read-only checks and return a combined status.
	status=0

	log_info "Starte Vorabprüfungen."

	verify_debian_version || status=1
	verify_root_rights || status=1
	verify_network_connection || status=1
	verify_free_diskspace || status=1
	verify_cpu_architecture || status=1
	verify_supported_board || status=1

	if [ "$status" -eq 0 ]; then
		log_success "Alle Vorabprüfungen erfolgreich."
	else
		log_error "Mindestens eine Vorabprüfung ist fehlgeschlagen."
	fi

	return "$status"
}

main() {
	# Entry point for manual verification runs.
	run_preflight_checks
}

if [ "${0##*/}" = "verify.sh" ]; then
	main "$@"
fi
