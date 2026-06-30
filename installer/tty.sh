#!/bin/sh
#
# tty autologin setup for kiosk-client Appliance Edition.
#
# Purpose:
#   Configures systemd getty autologin on tty1 for the kiosk user. This module
#   does not touch graphical login manager or session configuration.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
GETTY_OVERRIDE_DIR="/etc/systemd/system/getty@tty1.service.d"
GETTY_OVERRIDE_FILE="$GETTY_OVERRIDE_DIR/kiosk-autologin.conf"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

detect_kiosk_user() {
	# Determine the user account that should be logged in on tty1.
	if [ "${KIOSK_USER:-}" != "" ]; then
		printf '%s\n' "$KIOSK_USER"
		return 0
	fi

	if [ "${SUDO_USER:-}" != "" ] && [ "$SUDO_USER" != "root" ]; then
		printf '%s\n' "$SUDO_USER"
		return 0
	fi

	log_error "Kiosk-Benutzer konnte nicht ermittelt werden. Bitte KIOSK_USER setzen oder per sudo starten."
	return 1
}

validate_kiosk_user() {
	# Ensure the autologin user exists before writing the getty override.
	kiosk_user=$1

	if id "$kiosk_user" >/dev/null 2>&1; then
		return 0
	fi

	log_error "Kiosk-Benutzer existiert nicht: $kiosk_user"
	return 1
}

write_if_changed() {
	# Replace a file only when the generated content differs.
	target=$1
	temp_file=$2

	if [ -f "$target" ] && cmp -s "$temp_file" "$target"; then
		rm -f "$temp_file"
		log_success "Unveraendert: $target"
		return 0
	fi

	if ! mv "$temp_file" "$target"; then
		rm -f "$temp_file"
		log_error "Datei konnte nicht geschrieben werden: $target"
		return 1
	fi

	log_success "Aktualisiert: $target"
	return 0
}

configure_tty_autologin() {
	# Configure getty@tty1 to automatically log in the kiosk user.
	if ! require_root; then
		return 1
	fi

	kiosk_user=$(detect_kiosk_user)
	validate_kiosk_user "$kiosk_user"

	if ! mkdir -p "$GETTY_OVERRIDE_DIR"; then
		log_error "getty Override-Verzeichnis konnte nicht erstellt werden: $GETTY_OVERRIDE_DIR"
		return 1
	fi

	temp_file=$GETTY_OVERRIDE_FILE.tmp.$$
	cat >"$temp_file" <<EOF
[Service]
ExecStart=
ExecStart=-/sbin/agetty --autologin $kiosk_user --noclear %I \$TERM
EOF

	if ! write_if_changed "$GETTY_OVERRIDE_FILE" "$temp_file"; then
		return 1
	fi

	if ! chmod 0644 "$GETTY_OVERRIDE_FILE"; then
		log_error "Dateirechte konnten nicht gesetzt werden: $GETTY_OVERRIDE_FILE"
		return 1
	fi

	if command -v systemctl >/dev/null 2>&1; then
		systemctl daemon-reload
		systemctl enable getty@tty1.service
	fi

	log_success "tty1 Autologin fuer Benutzer '$kiosk_user' konfiguriert."
}

main() {
	configure_tty_autologin
}

if [ "${0##*/}" = "tty.sh" ]; then
	main "$@"
fi
