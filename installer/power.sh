#!/bin/sh
#
# Kiosk runtime power hardening for kiosk-client.
#
# Purpose:
#   Disables lock screen, screensaver, dimming, display power management, and
#   automatic sleep states for GNOME/GDM3 on Debian Bookworm. This module uses
#   gsettings and systemd configuration only. It does not use xset or shell
#   display hacks.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
LOGIND_DROPIN_DIR="/etc/systemd/logind.conf.d"
LOGIND_DROPIN="$LOGIND_DROPIN_DIR/kiosk-client.conf"
SLEEP_DROPIN_DIR="/etc/systemd/sleep.conf.d"
SLEEP_DROPIN="$SLEEP_DROPIN_DIR/kiosk-client.conf"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

detect_target_user() {
	# Determine the graphical user whose GNOME settings should be hardened.
	if [ "${KIOSK_USER:-}" != "" ]; then
		printf '%s\n' "$KIOSK_USER"
		return 0
	fi

	if [ "${SUDO_USER:-}" != "" ] && [ "$SUDO_USER" != "root" ]; then
		printf '%s\n' "$SUDO_USER"
		return 0
	fi

	log_error "Zielbenutzer konnte nicht ermittelt werden. Bitte KIOSK_USER setzen oder per sudo starten."
	return 1
}

get_user_uid() {
	# Resolve the target user's numeric uid.
	target_user=$1

	if user_uid=$(id -u "$target_user" 2>/dev/null); then
		printf '%s\n' "$user_uid"
		return 0
	fi

	log_error "UID für Benutzer '$target_user' konnte nicht ermittelt werden."
	return 1
}

run_user_gsettings() {
	# Apply gsettings for the target user. If a graphical user session is active,
	# use its runtime directory. Otherwise use a temporary D-Bus session so the
	# user's dconf database can still be updated during installation.
	target_user=$1
	user_uid=$2
	shift 2

	if [ -d "/run/user/$user_uid" ]; then
		sudo -u "$target_user" XDG_RUNTIME_DIR="/run/user/$user_uid" gsettings "$@"
		return $?
	fi

	if command -v dbus-run-session >/dev/null 2>&1; then
		sudo -u "$target_user" dbus-run-session -- gsettings "$@"
		return $?
	fi

	log_warn "Keine aktive User-Session und dbus-run-session nicht verfügbar; GNOME Einstellung wird übersprungen."
	return 2
}

set_user_gsetting() {
	# Set one GNOME setting and log a useful error if it fails.
	target_user=$1
	user_uid=$2
	schema=$3
	key=$4
	value=$5

	if ! writable=$(run_user_gsettings "$target_user" "$user_uid" writable "$schema" "$key" 2>/dev/null); then
		log_warn "GNOME Einstellung nicht verfügbar: $schema $key"
		return 0
	fi

	if [ "$writable" != "true" ]; then
		log_warn "GNOME Einstellung ist nicht schreibbar: $schema $key"
		return 0
	fi

	log_info "Setze GNOME Einstellung: $schema $key $value"
	if run_user_gsettings "$target_user" "$user_uid" set "$schema" "$key" "$value"; then
		return 0
	fi

	log_error "GNOME Einstellung konnte nicht gesetzt werden: $schema $key"
	return 1
}

configure_gnome_power_settings() {
	# Disable GNOME lock, screensaver activation, dimming, blanking, and suspend
	# behavior for the kiosk user.
	target_user=$1
	user_uid=$2

	if ! command -v gsettings >/dev/null 2>&1; then
		log_error "gsettings ist nicht verfügbar."
		return 1
	fi

	set_user_gsetting "$target_user" "$user_uid" org.gnome.desktop.session idle-delay "uint32 0"
	set_user_gsetting "$target_user" "$user_uid" org.gnome.desktop.screensaver lock-enabled false
	set_user_gsetting "$target_user" "$user_uid" org.gnome.desktop.screensaver idle-activation-enabled false
	set_user_gsetting "$target_user" "$user_uid" org.gnome.settings-daemon.plugins.power idle-dim false
	set_user_gsetting "$target_user" "$user_uid" org.gnome.settings-daemon.plugins.power sleep-inactive-ac-type "'nothing'"
	set_user_gsetting "$target_user" "$user_uid" org.gnome.settings-daemon.plugins.power sleep-inactive-battery-type "'nothing'"
}

write_logind_config() {
	# Configure systemd-logind to ignore idle actions and lid switches.
	if ! mkdir -p "$LOGIND_DROPIN_DIR"; then
		log_error "logind Drop-in-Verzeichnis konnte nicht erstellt werden: $LOGIND_DROPIN_DIR"
		return 1
	fi

	log_info "Schreibe systemd-logind Konfiguration: $LOGIND_DROPIN"
	if cat >"$LOGIND_DROPIN" <<'EOF'
[Login]
IdleAction=ignore
HandleLidSwitch=ignore
HandleLidSwitchExternalPower=ignore
HandleLidSwitchDocked=ignore
EOF
	then
		log_success "systemd-logind Konfiguration wurde geschrieben."
		return 0
	fi

	log_error "systemd-logind Konfiguration konnte nicht geschrieben werden."
	return 1
}

write_sleep_config() {
	# Disable systemd sleep states through the official sleep configuration.
	if ! mkdir -p "$SLEEP_DROPIN_DIR"; then
		log_error "systemd Sleep Drop-in-Verzeichnis konnte nicht erstellt werden: $SLEEP_DROPIN_DIR"
		return 1
	fi

	log_info "Schreibe systemd Sleep-Konfiguration: $SLEEP_DROPIN"
	if cat >"$SLEEP_DROPIN" <<'EOF'
[Sleep]
AllowSuspend=no
AllowHibernation=no
AllowHybridSleep=no
AllowSuspendThenHibernate=no
EOF
	then
		log_success "systemd Sleep-Konfiguration wurde geschrieben."
		return 0
	fi

	log_error "systemd Sleep-Konfiguration konnte nicht geschrieben werden."
	return 1
}

reload_logind() {
	# Reload systemd manager configuration and logind so drop-ins are noticed.
	log_info "Lade systemd Konfiguration neu."
	if ! systemctl daemon-reload; then
		log_error "systemctl daemon-reload fehlgeschlagen."
		return 1
	fi

	log_info "Starte systemd-logind neu."
	if ! systemctl restart systemd-logind.service; then
		log_error "systemd-logind konnte nicht neu gestartet werden."
		return 1
	fi
}

configure_power_hardening() {
	# Apply all runtime hardening settings.
	if ! require_root; then
		return 1
	fi

	target_user=$(detect_target_user)
	user_uid=$(get_user_uid "$target_user")

	log_info "Konfiguriere Kiosk Runtime Hardening für Benutzer: $target_user"

	configure_gnome_power_settings "$target_user" "$user_uid"
	write_logind_config
	write_sleep_config
	reload_logind

	log_success "Kiosk Runtime Hardening wurde konfiguriert."
}

main() {
	configure_power_hardening
}

if [ "${0##*/}" = "power.sh" ]; then
	main "$@"
fi
