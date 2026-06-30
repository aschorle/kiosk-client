#!/bin/sh
#
# systemd setup module for kiosk-client.
#
# Purpose:
#   Installs and enables the first productive kiosk-client systemd service:
#   kiosk-browser.service. This module only handles the browser autostart unit.
#   It does not configure Cage, Wayland, watchdogs, or a local web interface.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
SERVICE_NAME="kiosk-browser.service"
SERVICE_TEMPLATE="$PROJECT_DIR/systemd/$SERVICE_NAME"
SERVICE_TARGET="/etc/systemd/system/$SERVICE_NAME"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

detect_service_user() {
	# Determine the user account that should run Chromium.
	# KIOSK_USER can be set explicitly. During sudo-based installs, SUDO_USER is
	# the preferred default because it points to the login user, not root.
	if [ "${KIOSK_USER:-}" != "" ]; then
		printf '%s\n' "$KIOSK_USER"
		return 0
	fi

	if [ "${SUDO_USER:-}" != "" ] && [ "$SUDO_USER" != "root" ]; then
		printf '%s\n' "$SUDO_USER"
		return 0
	fi

	id -un
}

validate_service_user() {
	# Ensure the selected user exists before installing the system service.
	service_user=$1

	if id "$service_user" >/dev/null 2>&1; then
		return 0
	fi

	log_error "Systembenutzer existiert nicht: $service_user"
	return 1
}

escape_replacement() {
	# Escape values for use as awk replacement text.
	printf '%s\n' "$1" | awk '{ gsub(/\\/, "\\\\"); gsub(/&/, "\\&"); print }'
}

render_service_template() {
	# Render the repository service template with the detected user and path.
	service_user=$1
	escaped_user=$(escape_replacement "$service_user")
	escaped_project_dir=$(escape_replacement "$PROJECT_DIR")

	awk \
		-v user="$escaped_user" \
		-v project_dir="$escaped_project_dir" \
		'{
			gsub(/@KIOSK_USER@/, user)
			gsub(/@PROJECT_DIR@/, project_dir)
			print
		}' "$SERVICE_TEMPLATE"
}

install_browser_service() {
	# Install, enable, start, and verify kiosk-browser.service.
	if ! require_root; then
		return 1
	fi

	if [ ! -r "$SERVICE_TEMPLATE" ]; then
		log_error "Service-Vorlage nicht lesbar: $SERVICE_TEMPLATE"
		return 1
	fi

	service_user=$(detect_service_user)
	if ! validate_service_user "$service_user"; then
		return 1
	fi

	log_info "Installiere $SERVICE_NAME für Benutzer: $service_user"
	if render_service_template "$service_user" >"$SERVICE_TARGET"; then
		log_success "$SERVICE_NAME wurde nach $SERVICE_TARGET installiert."
	else
		log_error "$SERVICE_NAME konnte nicht nach $SERVICE_TARGET geschrieben werden."
		return 1
	fi

	log_info "Lade systemd-Konfiguration neu."
	if ! systemctl daemon-reload; then
		log_error "systemctl daemon-reload fehlgeschlagen."
		return 1
	fi

	log_info "Aktiviere $SERVICE_NAME."
	if ! systemctl enable "$SERVICE_NAME"; then
		log_error "$SERVICE_NAME konnte nicht aktiviert werden."
		return 1
	fi

	log_info "Starte $SERVICE_NAME."
	if ! systemctl restart "$SERVICE_NAME"; then
		log_error "$SERVICE_NAME konnte nicht gestartet werden."
		return 1
	fi

	log_info "Prüfe Status von $SERVICE_NAME."
	if systemctl is-active --quiet "$SERVICE_NAME"; then
		log_success "$SERVICE_NAME läuft."
		return 0
	fi

	systemctl status "$SERVICE_NAME" --no-pager || true
	log_error "$SERVICE_NAME ist nicht aktiv."
	return 1
}

main() {
	install_browser_service
}

if [ "${0##*/}" = "systemd.sh" ]; then
	main "$@"
fi
