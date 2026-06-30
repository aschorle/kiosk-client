#!/bin/sh
#
# Appliance Edition installer for kiosk-client.
#
# Purpose:
#   Installs the minimal kiosk-client appliance profile without a display
#   manager or desktop environment. The existing Desktop Edition installer is
#   not used or modified by this profile.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
APPLIANCE_PACKAGES="
chromium
cage
dbus
"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

install_appliance_packages() {
	# Install only the package set required by the Appliance Edition runtime.
	DEBIAN_FRONTEND=noninteractive
	export DEBIAN_FRONTEND

	set --
	for package in $APPLIANCE_PACKAGES; do
		set -- "$@" "$package"
	done

	log_info "Installiere Appliance-Pakete: $*"
	apt install -y --no-install-recommends "$@"
}

run_module() {
	# Execute one appliance installer module.
	module_name=$1
	module_path=$SCRIPT_DIR/$module_name

	if [ ! -r "$module_path" ]; then
		log_error "Appliance-Modul nicht lesbar: $module_path"
		return 1
	fi

	sh "$module_path"
}

install_appliance() {
	# Install the minimal appliance profile in a deterministic order.
	require_root
	check_debian
	check_architecture
	check_network
	install_appliance_packages
	run_module "runtime.sh"
	run_module "tty.sh"
	log_success "Appliance Edition Installation abgeschlossen."
}

main() {
	install_appliance
}

if [ "${0##*/}" = "appliance.sh" ]; then
	main "$@"
fi
