#!/bin/sh
#
# Browser installation module for kiosk-client.
#
# Purpose:
#   Installs and verifies Chromium as the browser runtime used by later kiosk
#   phases. This module does not configure kiosk mode, browser flags, policies,
#   cache handling, autostart, Wayland, Cage, systemd units, or URL settings.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

CHROMIUM_PACKAGE="chromium"
CHROMIUM_PATH=""
CHROMIUM_VERSION=""

find_chromium_binary() {
	# Resolve the Chromium executable path.
	# Debian 12 provides the binary as "chromium"; the fallback keeps the check
	# tolerant for images that still expose "chromium-browser".
	if command -v chromium >/dev/null 2>&1; then
		command -v chromium
		return 0
	fi

	if command -v chromium-browser >/dev/null 2>&1; then
		command -v chromium-browser
		return 0
	fi

	return 1
}

install_chromium() {
	# Install Chromium through apt.
	# The function is idempotent: if a Chromium executable is already available,
	# no package installation is attempted.
	if CHROMIUM_PATH="$(find_chromium_binary)"; then
		log_success "Chromium ist bereits installiert: $CHROMIUM_PATH."
		return 0
	fi

	if ! require_root; then
		return 1
	fi

	log_info "Installiere Chromium über apt."

	DEBIAN_FRONTEND=noninteractive
	export DEBIAN_FRONTEND

	if apt install -y --no-install-recommends "$CHROMIUM_PACKAGE"; then
		log_success "Chromium wurde installiert."
		return 0
	fi

	log_error "Chromium konnte nicht installiert werden."
	return 1
}

verify_chromium() {
	# Verify that Chromium is installed and collect executable path and version.
	# The values are stored in CHROMIUM_PATH and CHROMIUM_VERSION for the summary.
	if ! CHROMIUM_PATH="$(find_chromium_binary)"; then
		log_error "Chromium wurde nicht gefunden."
		return 1
	fi

	if ! CHROMIUM_VERSION="$("$CHROMIUM_PATH" --version 2>/dev/null)"; then
		log_error "Chromium-Version konnte nicht ermittelt werden."
		return 1
	fi

	log_success "Chromium gefunden: $CHROMIUM_PATH."
	log_info "Chromium-Version: $CHROMIUM_VERSION."
	return 0
}

show_browser_summary() {
	# Print the browser installation summary.
	# If the values are not populated yet, run verification first.
	if [ "$CHROMIUM_PATH" = "" ] || [ "$CHROMIUM_VERSION" = "" ]; then
		if ! verify_chromium; then
			return 1
		fi
	fi

	log_info "Browser-Zusammenfassung:"
	log_info "Installationspfad: $CHROMIUM_PATH"
	log_info "Versionsnummer: $CHROMIUM_VERSION"
}

main() {
	# Direct execution installs Chromium, verifies the result, and prints the
	# resolved path and version. No kiosk behavior is enabled here.
	install_chromium
	verify_chromium
	show_browser_summary
}

if [ "${0##*/}" = "browser.sh" ]; then
	main "$@"
fi
