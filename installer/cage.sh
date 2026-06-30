#!/bin/sh
#
# Cage installation module for kiosk-client.
#
# Purpose:
#   Installs the minimal package base required for running Cage later. This
#   module does not configure Cage sessions, change boot behavior, remove a
#   desktop environment, or start Chromium inside Cage yet.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
CAGE_PACKAGES="
cage
"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

install_cage_packages() {
	# Install Cage package dependencies. apt is idempotent, so re-running this
	# module keeps already installed packages unchanged.
	if ! require_root; then
		return 1
	fi

	set --
	for package in $CAGE_PACKAGES; do
		set -- "$@" "$package"
	done

	if [ "$#" -eq 0 ]; then
		log_warn "Keine Cage-Pakete definiert."
		return 0
	fi

	log_info "Installiere Cage-Pakete."

	DEBIAN_FRONTEND=noninteractive
	export DEBIAN_FRONTEND

	if apt install -y --no-install-recommends "$@"; then
		log_success "Cage-Pakete wurden installiert."
		return 0
	fi

	log_error "Cage-Pakete konnten nicht installiert werden."
	return 1
}

verify_cage() {
	# Verify that Cage is installed and print the detected version.
	if ! cage_path=$(command -v cage 2>/dev/null); then
		log_error "Cage wurde nicht gefunden. Bitte installer/cage.sh erneut ausführen."
		return 1
	fi

	log_success "Cage gefunden: $cage_path"

	if cage_version=$(cage --version 2>&1); then
		log_info "Cage Version: $cage_version"
		return 0
	fi

	log_warn "Cage Version konnte nicht ermittelt werden."
	return 0
}

main() {
	install_cage_packages
	verify_cage
}

if [ "${0##*/}" = "cage.sh" ]; then
	main "$@"
fi
