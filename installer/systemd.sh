#!/bin/sh
#
# systemd user-service setup module for kiosk-client.
#
# Purpose:
#   Installs and enables kiosk-browser.service as a systemd user service. The
#   browser must run inside the graphical user session, never as root and never
#   as a system service.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
SERVICE_NAME="kiosk-browser.service"
SERVICE_SOURCE="$PROJECT_DIR/systemd/user/$SERVICE_NAME"
USER_SYSTEMD_DIR="$HOME/.config/systemd/user"
SERVICE_TARGET="$USER_SYSTEMD_DIR/$SERVICE_NAME"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

require_user_context() {
	# User services must be installed from the target login user session.
	if [ "$(id -u)" -eq 0 ]; then
		log_error "User-Service darf nicht als root installiert werden."
		log_error "Bitte als grafischer Zielbenutzer ausführen: ./installer/systemd.sh"
		return 1
	fi

	return 0
}

install_browser_service() {
	# Install, enable, restart, and verify the systemd user service.
	if ! require_user_context; then
		return 1
	fi

	if [ ! -r "$SERVICE_SOURCE" ]; then
		log_error "Service-Datei nicht lesbar: $SERVICE_SOURCE"
		return 1
	fi

	log_info "Installiere User-Service nach $SERVICE_TARGET."
	if ! mkdir -p "$USER_SYSTEMD_DIR"; then
		log_error "User-systemd-Verzeichnis konnte nicht erstellt werden: $USER_SYSTEMD_DIR"
		return 1
	fi

	if ! cp "$SERVICE_SOURCE" "$SERVICE_TARGET"; then
		log_error "$SERVICE_NAME konnte nicht installiert werden."
		return 1
	fi

	log_info "Lade systemd User-Konfiguration neu."
	if ! systemctl --user daemon-reload; then
		log_error "systemctl --user daemon-reload fehlgeschlagen."
		return 1
	fi

	log_info "Aktiviere $SERVICE_NAME als User-Service."
	if ! systemctl --user enable "$SERVICE_NAME"; then
		log_error "$SERVICE_NAME konnte nicht aktiviert werden."
		return 1
	fi

	log_info "Starte $SERVICE_NAME neu."
	if ! systemctl --user restart "$SERVICE_NAME"; then
		log_error "$SERVICE_NAME konnte nicht gestartet werden."
		return 1
	fi

	log_info "Prüfe Status von $SERVICE_NAME."
	if systemctl --user is-active --quiet "$SERVICE_NAME"; then
		log_success "$SERVICE_NAME läuft als User-Service."
		return 0
	fi

	systemctl --user status "$SERVICE_NAME" --no-pager || true
	log_error "$SERVICE_NAME ist nicht aktiv."
	return 1
}

main() {
	install_browser_service
}

if [ "${0##*/}" = "systemd.sh" ]; then
	main "$@"
fi
