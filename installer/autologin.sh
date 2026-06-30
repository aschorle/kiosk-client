#!/bin/sh
#
# GDM3 autologin setup for kiosk-client.
#
# Purpose:
#   Enables automatic login for the kiosk user in GDM3 so the graphical user
#   session starts after boot and the systemd user service can launch Chromium.
#   Only GDM3 is supported by this module.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
GDM3_CONFIG="/etc/gdm3/daemon.conf"
GDM3_BACKUP="/etc/gdm3/daemon.conf.bak"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

detect_autologin_user() {
	# Determine the user account that should be logged in automatically.
	if [ "${KIOSK_USER:-}" != "" ]; then
		printf '%s\n' "$KIOSK_USER"
		return 0
	fi

	if [ "${SUDO_USER:-}" != "" ] && [ "$SUDO_USER" != "root" ]; then
		printf '%s\n' "$SUDO_USER"
		return 0
	fi

	log_error "Autologin-Benutzer konnte nicht ermittelt werden. Bitte KIOSK_USER setzen oder per sudo starten."
	return 1
}

validate_autologin_user() {
	# Ensure the configured autologin user exists.
	autologin_user=$1

	if id "$autologin_user" >/dev/null 2>&1; then
		return 0
	fi

	log_error "Autologin-Benutzer existiert nicht: $autologin_user"
	return 1
}

ensure_gdm3_config_exists() {
	# GDM3 uses /etc/gdm3/daemon.conf. If the file does not exist yet, create it
	# with a daemon section so the later update remains simple and deterministic.
	if [ -e "$GDM3_CONFIG" ]; then
		return 0
	fi

	log_warn "$GDM3_CONFIG existiert nicht; neue GDM3-Konfiguration wird angelegt."
	if ! printf '%s\n' "[daemon]" >"$GDM3_CONFIG"; then
		log_error "$GDM3_CONFIG konnte nicht angelegt werden."
		return 1
	fi
}

backup_gdm3_config() {
	# Create the required backup before modifying daemon.conf.
	if ! cp "$GDM3_CONFIG" "$GDM3_BACKUP"; then
		log_error "Sicherung konnte nicht erstellt werden: $GDM3_BACKUP"
		return 1
	fi

	log_success "Sicherung erstellt: $GDM3_BACKUP"
}

is_autologin_configured() {
	# Check whether daemon.conf already contains the desired autologin settings.
	autologin_user=$1

	awk -v user="$autologin_user" '
		BEGIN {
			in_daemon = 0
			enable_ok = 0
			user_ok = 0
		}
		/^[[:space:]]*\[/ {
			in_daemon = ($0 ~ /^[[:space:]]*\[daemon\][[:space:]]*$/)
			next
		}
		in_daemon && /^[[:space:]]*AutomaticLoginEnable[[:space:]]*=/ {
			value = $0
			sub(/^[^=]*=/, "", value)
			gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
			if (value == "True") {
				enable_ok = 1
			}
		}
		in_daemon && /^[[:space:]]*AutomaticLogin[[:space:]]*=/ {
			value = $0
			sub(/^[^=]*=/, "", value)
			gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
			if (value == user) {
				user_ok = 1
			}
		}
		END {
			exit !(enable_ok && user_ok)
		}
	' "$GDM3_CONFIG"
}

write_gdm3_autologin_config() {
	# Update or add AutomaticLoginEnable and AutomaticLogin inside [daemon].
	autologin_user=$1
	temp_config="${GDM3_CONFIG}.tmp.$$"

	if ! awk -v user="$autologin_user" '
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
		{
			print
		}
		END {
			write_missing()
			if (!seen_daemon) {
				print ""
				print "[daemon]"
				print "AutomaticLoginEnable=True"
				print "AutomaticLogin=" user
			}
		}
	' "$GDM3_CONFIG" >"$temp_config"; then
		log_error "GDM3-Konfiguration konnte nicht vorbereitet werden."
		return 1
	fi

	if ! mv "$temp_config" "$GDM3_CONFIG"; then
		log_error "GDM3-Konfiguration konnte nicht geschrieben werden."
		return 1
	fi
}

configure_gdm3_autologin() {
	# Enable GDM3 autologin for the detected kiosk user.
	if ! require_root; then
		return 1
	fi

	autologin_user=$(detect_autologin_user)
	if ! validate_autologin_user "$autologin_user"; then
		return 1
	fi

	log_info "Konfiguriere GDM3 Autologin für Benutzer: $autologin_user"

	if ! ensure_gdm3_config_exists; then
		return 1
	fi

	if is_autologin_configured "$autologin_user"; then
		log_success "GDM3 Autologin ist bereits korrekt konfiguriert."
		return 0
	fi

	if ! backup_gdm3_config; then
		return 1
	fi

	if ! write_gdm3_autologin_config "$autologin_user"; then
		return 1
	fi

	log_success "GDM3 Autologin wurde konfiguriert."
}

main() {
	configure_gdm3_autologin
}

if [ "${0##*/}" = "autologin.sh" ]; then
	main "$@"
fi
