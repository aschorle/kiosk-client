#!/bin/sh
#
# Native kiosk session setup for kiosk-client.
#
# Purpose:
#   Installs a dedicated display-manager session named "kiosk" which starts
#   scripts/start-cage.sh directly. GNOME/KDE stay installed as fallback, but
#   autologin is switched to the native Cage kiosk session.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
SESSION_NAME="kiosk"
SESSION_FILE="$SESSION_NAME.desktop"
WAYLAND_SESSION_DIR="/usr/share/wayland-sessions"
LEGACY_XSESSION_FILE="/usr/share/xsessions/$SESSION_FILE"
GDM3_CONFIG="/etc/gdm3/daemon.conf"
ACCOUNTS_SERVICE_DIR="/var/lib/AccountsService/users"
SDDM_CONFIG_DIR="/etc/sddm.conf.d"
SDDM_KIOSK_CONFIG="$SDDM_CONFIG_DIR/kiosk-client.conf"
LIGHTDM_CONFIG_DIR="/etc/lightdm/lightdm.conf.d"
LIGHTDM_KIOSK_CONFIG="$LIGHTDM_CONFIG_DIR/50-kiosk-client.conf"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

detect_session_user() {
	# Determine the user account that should receive autologin.
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

validate_session_user() {
	# Ensure that the target kiosk user exists.
	session_user=$1

	if id "$session_user" >/dev/null 2>&1; then
		return 0
	fi

	log_error "Kiosk-Benutzer existiert nicht: $session_user"
	return 1
}

detect_display_manager() {
	# Prefer the system default display-manager link, then fall back to known
	# configuration directories and installed binaries.
	default_dm=""
	if [ -r /etc/X11/default-display-manager ]; then
		read -r default_dm_path </etc/X11/default-display-manager || default_dm_path=""
		default_dm=$(basename "$default_dm_path")
	fi

	case "$default_dm" in
		gdm3|gdm)
			printf '%s\n' "gdm"
			return 0
			;;
		sddm)
			printf '%s\n' "sddm"
			return 0
			;;
		lightdm)
			printf '%s\n' "lightdm"
			return 0
			;;
	esac

	if command -v gdm3 >/dev/null 2>&1 || command -v gdm >/dev/null 2>&1 || [ -d /etc/gdm3 ]; then
		printf '%s\n' "gdm"
		return 0
	fi

	if command -v sddm >/dev/null 2>&1 || [ -d /etc/sddm.conf.d ] || [ -f /etc/sddm.conf ]; then
		printf '%s\n' "sddm"
		return 0
	fi

	if command -v lightdm >/dev/null 2>&1 || [ -d /etc/lightdm ]; then
		printf '%s\n' "lightdm"
		return 0
	fi

	log_error "Kein unterstuetzter Display Manager gefunden. Unterstuetzt: GDM, SDDM, LightDM."
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
		log_error "Datei konnte nicht geschrieben werden: $target"
		rm -f "$temp_file"
		return 1
	fi

	log_success "Aktualisiert: $target"
	return 0
}

write_session_file() {
	# Write one display-manager session file.
	session_target=$1
	session_exec=$2
	temp_file=$session_target.tmp.$$

	cat >"$temp_file" <<EOF
[Desktop Entry]
Name=kiosk
Comment=kiosk-client native Cage session
Exec=$session_exec
TryExec=$session_exec
Type=Application
DesktopNames=kiosk
X-GDM-SessionRegisters=true
EOF

	if ! write_if_changed "$session_target" "$temp_file"; then
		return 1
	fi

	if ! chmod 0644 "$session_target"; then
		log_error "Dateirechte konnten nicht gesetzt werden: $session_target"
		return 1
	fi
}

remove_legacy_xsession_file() {
	# Older kiosk-client installers registered kiosk as an XSession. That can
	# make GDM start the session via its X11 path. Remove only the kiosk-client
	# owned legacy file and leave unrelated sessions untouched.
	if [ ! -e "$LEGACY_XSESSION_FILE" ]; then
		return 0
	fi

	if ! grep -q "kiosk-client native Cage session" "$LEGACY_XSESSION_FILE"; then
		log_warn "Legacy XSession-Datei gehoert nicht eindeutig kiosk-client und bleibt erhalten: $LEGACY_XSESSION_FILE"
		return 0
	fi

	if rm "$LEGACY_XSESSION_FILE"; then
		log_success "Legacy XSession entfernt: $LEGACY_XSESSION_FILE"
		return 0
	fi

	log_error "Legacy XSession konnte nicht entfernt werden: $LEGACY_XSESSION_FILE"
	return 1
}

install_kiosk_session_file() {
	# Install the native Wayland session used by the display manager. Do not
	# register this as an XSession; otherwise GDM may start the kiosk path through
	# its X11 session wrapper and leave the native Cage session non-exclusive.
	session_user=$1
	session_exec=$PROJECT_DIR/scripts/start-cage.sh
	wayland_session_target=$WAYLAND_SESSION_DIR/$SESSION_FILE

	if [ ! -x "$session_exec" ]; then
		log_error "Cage-Startskript ist nicht ausfuehrbar: $session_exec"
		return 1
	fi

	if ! mkdir -p "$WAYLAND_SESSION_DIR"; then
		log_error "Session-Verzeichnis konnte nicht erstellt werden: $WAYLAND_SESSION_DIR"
		return 1
	fi

	if ! write_session_file "$wayland_session_target" "$session_exec"; then
		return 1
	fi

	if ! remove_legacy_xsession_file; then
		return 1
	fi

	log_info "Session '$SESSION_NAME' fuer Benutzer '$session_user' installiert."
}

ensure_gdm_config_exists() {
	if [ -e "$GDM3_CONFIG" ]; then
		return 0
	fi

	if ! mkdir -p "$(dirname "$GDM3_CONFIG")"; then
		log_error "GDM-Konfigurationsverzeichnis konnte nicht erstellt werden."
		return 1
	fi

	printf '%s\n' "[daemon]" >"$GDM3_CONFIG"
}

configure_gdm_autologin() {
	# Configure GDM autologin and set the user session via AccountsService.
	session_user=$1
	temp_file=$GDM3_CONFIG.tmp.$$

	if ! ensure_gdm_config_exists; then
		return 1
	fi

	if ! awk -v user="$session_user" '
		BEGIN {
			in_daemon = 0
			seen_daemon = 0
			wrote_enable = 0
			wrote_user = 0
		}
		function write_missing() {
			if (in_daemon && !wrote_enable) {
				print "AutomaticLoginEnable=True"
				wrote_enable = 1
			}
			if (in_daemon && !wrote_user) {
				print "AutomaticLogin=" user
				wrote_user = 1
			}
		}
		/^[[:space:]]*\[/ {
			write_missing()
			in_daemon = ($0 ~ /^[[:space:]]*\[daemon\][[:space:]]*$/)
			if (in_daemon) {
				seen_daemon = 1
			}
			print
			next
		}
		in_daemon && /^[[:space:]]*AutomaticLoginEnable[[:space:]]*=/ {
			if (!wrote_enable) {
				print "AutomaticLoginEnable=True"
				wrote_enable = 1
			}
			next
		}
		in_daemon && /^[[:space:]]*AutomaticLogin[[:space:]]*=/ {
			if (!wrote_user) {
				print "AutomaticLogin=" user
				wrote_user = 1
			}
			next
		}
		{ print }
		END {
			write_missing()
			if (!seen_daemon) {
				print ""
				print "[daemon]"
				print "AutomaticLoginEnable=True"
				print "AutomaticLogin=" user
			}
		}
	' "$GDM3_CONFIG" >"$temp_file"; then
		log_error "GDM-Konfiguration konnte nicht vorbereitet werden."
		rm -f "$temp_file"
		return 1
	fi

	if ! write_if_changed "$GDM3_CONFIG" "$temp_file"; then
		return 1
	fi

	configure_accountsservice_session "$session_user"
	restart_accounts_daemon_if_available
}

configure_accountsservice_session() {
	# GDM reads the preferred autologin session from AccountsService.
	session_user=$1
	account_file=$ACCOUNTS_SERVICE_DIR/$session_user
	temp_file=$account_file.tmp.$$

	if ! mkdir -p "$ACCOUNTS_SERVICE_DIR"; then
		log_error "AccountsService-Verzeichnis konnte nicht erstellt werden: $ACCOUNTS_SERVICE_DIR"
		return 1
	fi

	if [ ! -f "$account_file" ]; then
		printf '%s\n' "[User]" >"$account_file"
	fi

	if ! awk -v session="$SESSION_NAME" '
		BEGIN {
			in_user = 0
			seen_user = 0
			wrote_session = 0
		}
		function write_missing() {
			if (in_user && !wrote_session) {
				print "Session=" session
				wrote_session = 1
			}
		}
		/^[[:space:]]*\[/ {
			write_missing()
			in_user = ($0 ~ /^[[:space:]]*\[User\][[:space:]]*$/)
			if (in_user) {
				seen_user = 1
			}
			print
			next
		}
		in_user && /^[[:space:]]*XSession[[:space:]]*=/ { next }
		in_user && /^[[:space:]]*Session[[:space:]]*=/ {
			if (!wrote_session) {
				print "Session=" session
				wrote_session = 1
			}
			next
		}
		{ print }
		END {
			write_missing()
			if (!seen_user) {
				print ""
				print "[User]"
				print "Session=" session
			}
		}
	' "$account_file" >"$temp_file"; then
		log_error "AccountsService-Konfiguration konnte nicht vorbereitet werden."
		rm -f "$temp_file"
		return 1
	fi

	if ! write_if_changed "$account_file" "$temp_file"; then
		return 1
	fi

	if ! chown root:root "$account_file"; then
		log_error "Besitzrechte konnten nicht gesetzt werden: $account_file"
		return 1
	fi

	if ! chmod 0644 "$account_file"; then
		log_error "Dateirechte konnten nicht gesetzt werden: $account_file"
		return 1
	fi

	log_success "GDM AccountsService session set to kiosk"
}

restart_accounts_daemon_if_available() {
	# GDM may cache AccountsService data. Restart accounts-daemon when the
	# service exists, but keep installation successful on images without it.
	if ! command -v systemctl >/dev/null 2>&1; then
		log_warn "systemctl nicht gefunden; accounts-daemon wurde nicht neu gestartet."
		return 0
	fi

	if ! systemctl cat accounts-daemon.service >/dev/null 2>&1; then
		log_warn "accounts-daemon.service nicht gefunden; Neustart uebersprungen."
		return 0
	fi

	if systemctl restart accounts-daemon.service; then
		log_success "accounts-daemon.service wurde neu gestartet."
		return 0
	fi

	log_warn "accounts-daemon.service konnte nicht neu gestartet werden; Installation wird fortgesetzt."
	return 0
}

configure_sddm_autologin() {
	# Configure SDDM autologin through a dedicated drop-in.
	session_user=$1
	temp_file=$SDDM_KIOSK_CONFIG.tmp.$$

	if ! mkdir -p "$SDDM_CONFIG_DIR"; then
		log_error "SDDM-Konfigurationsverzeichnis konnte nicht erstellt werden: $SDDM_CONFIG_DIR"
		return 1
	fi

	cat >"$temp_file" <<EOF
[Autologin]
User=$session_user
Session=$SESSION_FILE
Relogin=true
EOF

	if ! write_if_changed "$SDDM_KIOSK_CONFIG" "$temp_file"; then
		return 1
	fi

	chmod 0644 "$SDDM_KIOSK_CONFIG"
}

configure_lightdm_autologin() {
	# Configure LightDM autologin through a dedicated drop-in.
	session_user=$1
	temp_file=$LIGHTDM_KIOSK_CONFIG.tmp.$$

	if ! mkdir -p "$LIGHTDM_CONFIG_DIR"; then
		log_error "LightDM-Konfigurationsverzeichnis konnte nicht erstellt werden: $LIGHTDM_CONFIG_DIR"
		return 1
	fi

	cat >"$temp_file" <<EOF
[Seat:*]
autologin-user=$session_user
autologin-session=$SESSION_NAME
user-session=$SESSION_NAME
EOF

	if ! write_if_changed "$LIGHTDM_KIOSK_CONFIG" "$temp_file"; then
		return 1
	fi

	chmod 0644 "$LIGHTDM_KIOSK_CONFIG"
}

configure_display_manager_session() {
	# Apply the autologin session configuration for the detected display manager.
	display_manager=$1
	session_user=$2

	case "$display_manager" in
		gdm)
			configure_gdm_autologin "$session_user"
			;;
		sddm)
			configure_sddm_autologin "$session_user"
			;;
		lightdm)
			configure_lightdm_autologin "$session_user"
			;;
		*)
			log_error "Nicht unterstuetzter Display Manager: $display_manager"
			return 1
			;;
	esac
}

install_native_kiosk_session() {
	# Install the session and switch autologin to it.
	if ! require_root; then
		return 1
	fi

	session_user=$(detect_session_user)
	validate_session_user "$session_user"
	display_manager=$(detect_display_manager)

	log_info "Display Manager erkannt: $display_manager"
	log_info "Kiosk-Session-Benutzer: $session_user"

	install_kiosk_session_file "$session_user"
	configure_display_manager_session "$display_manager" "$session_user"

	log_success "Native Kiosk Session '$SESSION_NAME' wurde konfiguriert."
}

main() {
	install_native_kiosk_session
}

if [ "${0##*/}" = "session.sh" ]; then
	main "$@"
fi
