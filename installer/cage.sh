#!/bin/sh
#
# Cage installation module for kiosk-client.
#
# Purpose:
#   Installs and verifies Cage for the productive appliance runtime. This
#   module does not remove GNOME/GDM and does not switch display managers.

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
		log_error "Cage wurde nicht gefunden. Bitte pruefen, ob das Paket 'cage' installiert werden konnte."
		return 1
	fi

	log_success "Cage gefunden: $cage_path"

	if cage_version=$(cage -v 2>&1); then
		log_info "Cage Version: $cage_version"
		return 0
	fi

	log_error "Cage ist vorhanden, aber 'cage -v' konnte nicht ausgefuehrt werden."
	return 1
}

main() {
	install_cage_packages
	verify_cage
}

if [ "${0##*/}" = "cage.sh" ]; then
	main "$@"
fi
