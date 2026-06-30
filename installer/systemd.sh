#!/bin/sh
#
# systemd user-service setup module for kiosk-client.
#
# Purpose:
#   Installs the kiosk-client systemd user services. The productive appliance
#   runtime is kiosk-runtime.service. kiosk-browser.service is kept only as a
#   manual legacy/fallback service and is disabled during installation.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
ENABLED_USER_SERVICES="
kiosk-agent.service
kiosk-runtime.service
"
FALLBACK_USER_SERVICES="
kiosk-browser.service
"

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

detect_target_user() {
	# Select the user account that owns the systemd user service.
	if [ "${KIOSK_USER:-}" != "" ]; then
		printf '%s\n' "$KIOSK_USER"
		return 0
	fi

	if [ "${SUDO_USER:-}" != "" ]; then
		printf '%s\n' "$SUDO_USER"
		return 0
	fi

	log_error "Zielbenutzer konnte nicht ermittelt werden. Bitte KIOSK_USER setzen oder per sudo starten."
	return 1
}

get_user_home() {
	# Resolve the target user's home directory from the system account database.
	target_user=$1

	home_dir=$(getent passwd "$target_user" | awk -F ':' '{ print $6 }')

	if [ "$home_dir" != "" ]; then
		printf '%s\n' "$home_dir"
		return 0
	fi

	log_error "Home-Verzeichnis für Benutzer '$target_user' konnte nicht ermittelt werden."
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

run_user_systemctl() {
	# Run systemctl --user as the target user with the correct runtime directory.
	target_user=$1
	user_uid=$2
	shift 2

	sudo -u "$target_user" XDG_RUNTIME_DIR="/run/user/$user_uid" systemctl --user "$@"
}

get_service_wanted_by() {
	# Read the first WantedBy= target from the service's [Install] section.
	# This keeps manual fallback enabling aligned with systemctl --user enable.
	service_name=$1
	service_source=$PROJECT_DIR/systemd/user/$service_name

	wanted_by=$(awk -F '=' '
		/^[[:space:]]*\[Install\][[:space:]]*$/ {
			in_install = 1
			next
		}
		/^[[:space:]]*\[/ {
			in_install = 0
		}
		in_install && /^[[:space:]]*WantedBy[[:space:]]*=/ {
			value = $2
			gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
			split(value, targets, /[[:space:]]+/)
			print targets[1]
			exit
		}
	' "$service_source")

	if [ "$wanted_by" != "" ]; then
		printf '%s\n' "$wanted_by"
		return 0
	fi

	log_error "WantedBy= fehlt in Service-Datei: $service_source"
	return 1
}

is_user_manager_available() {
	# The systemd user manager is reachable only while /run/user/<uid> exists.
	user_uid=$1

	[ -d "/run/user/$user_uid" ]
}

is_graphical_session_active() {
	# Verify that the service is restarted only inside an active graphical
	# session. This avoids detached Chromium processes after manual installer
	# runs from SSH or a non-graphical context.
	target_user=$1
	user_uid=$2

	run_user_systemctl "$target_user" "$user_uid" is-active --quiet graphical-session.target
}

enable_user_service_without_session() {
	# Enable the user service by creating the WantedBy target link from the unit.
	# This is used when no active user session exists yet and systemctl --user
	# cannot talk to the user manager.
	target_user=$1
	user_home=$2
	service_name=$3

	if ! wanted_by=$(get_service_wanted_by "$service_name"); then
		return 1
	fi

	user_systemd_dir=$user_home/.config/systemd/user
	wants_dir=$user_systemd_dir/$wanted_by.wants
	service_target=$user_systemd_dir/$service_name
	wants_link=$wants_dir/$service_name

	if ! mkdir -p "$wants_dir"; then
		log_error "User-Service-Aktivierungsverzeichnis konnte nicht erstellt werden: $wants_dir"
		return 1
	fi

	if [ ! -e "$wants_link" ] && ! ln -s "../$service_name" "$wants_link"; then
		log_error "$service_name konnte nicht für $wanted_by aktiviert werden."
		return 1
	fi

	if ! chown -h "$target_user:$target_user" "$wants_dir" "$wants_link" "$service_target"; then
		log_error "Besitzrechte für die User-Service-Aktivierung konnten nicht gesetzt werden."
		return 1
	fi

	return 0
}

disable_user_service_without_session() {
	# Disable a user service by removing the symlink created below its
	# WantedBy= target. This is the offline equivalent of systemctl --user
	# disable when no user manager is reachable yet.
	target_user=$1
	user_home=$2
	service_name=$3

	if ! wanted_by=$(get_service_wanted_by "$service_name"); then
		return 1
	fi

	user_systemd_dir=$user_home/.config/systemd/user
	wants_link=$user_systemd_dir/$wanted_by.wants/$service_name

	if [ -L "$wants_link" ]; then
		log_info "Deaktiviere Fallback-Service-Link: $wants_link"
		if ! rm "$wants_link"; then
			log_error "$service_name konnte nicht deaktiviert werden."
			return 1
		fi
	elif [ -e "$wants_link" ]; then
		log_warn "Aktivierungseintrag ist kein Symlink und wurde nicht entfernt: $wants_link"
	fi

	return 0
}

remove_legacy_default_target_link() {
	# Older kiosk-client versions enabled kiosk-browser.service below
	# default.target. That can start Chromium before the graphical session is
	# ready. Remove only that legacy symlink and leave all other files untouched.
	target_user=$1
	user_home=$2
	service_name=$3
	legacy_link=$user_home/.config/systemd/user/default.target.wants/$service_name

	if [ "$service_name" != "kiosk-browser.service" ]; then
		return 0
	fi

	if [ -L "$legacy_link" ]; then
		log_info "Entferne alten default.target-Link: $legacy_link"
		if ! rm "$legacy_link"; then
			log_error "Alter default.target-Link konnte nicht entfernt werden: $legacy_link"
			return 1
		fi
		return 0
	fi

	if [ -e "$legacy_link" ]; then
		log_warn "Alter default.target-Eintrag ist kein Symlink und wurde nicht entfernt: $legacy_link"
	fi

	return 0
}

install_service_file() {
	# Copy the user service and assign ownership to the target user.
	target_user=$1
	user_home=$2
	service_name=$3
	service_source=$PROJECT_DIR/systemd/user/$service_name
	user_systemd_dir=$user_home/.config/systemd/user
	service_target=$user_systemd_dir/$service_name

	if [ ! -r "$service_source" ]; then
		log_error "Service-Datei nicht lesbar: $service_source"
		return 1
	fi

	log_info "Installing user service: $service_name"
	if ! mkdir -p "$user_systemd_dir"; then
		log_error "User-systemd-Verzeichnis konnte nicht erstellt werden: $user_systemd_dir"
		return 1
	fi

	if ! cp "$service_source" "$service_target"; then
		log_error "$service_name konnte nicht installiert werden."
		return 1
	fi

	if ! chown "$target_user:$target_user" "$user_systemd_dir" "$service_target"; then
		log_error "Besitzrechte für $service_name konnten nicht gesetzt werden."
		return 1
	fi

	if ! chmod 0644 "$service_target"; then
		log_error "Dateirechte für $service_name konnten nicht gesetzt werden."
		return 1
	fi

	log_success "$service_name installiert: $service_target"
	return 0
}

install_user_service() {
	# Install, enable, restart, and verify one systemd user service.
	target_user=$1
	user_home=$2
	user_uid=$3
	service_name=$4
	start_policy=${5:-start}

	if ! install_service_file "$target_user" "$user_home" "$service_name"; then
		return 1
	fi

	if ! remove_legacy_default_target_link "$target_user" "$user_home" "$service_name"; then
		return 1
	fi

	if is_user_manager_available "$user_uid"; then
		log_info "Reloading user daemon..."
		if ! run_user_systemctl "$target_user" "$user_uid" daemon-reload; then
			log_error "systemctl --user daemon-reload fehlgeschlagen."
			return 1
		fi

		log_info "Enabling service: $service_name"
		if ! run_user_systemctl "$target_user" "$user_uid" enable "$service_name"; then
			log_error "$service_name konnte nicht aktiviert werden."
			return 1
		fi
	else
		log_info "Enabling service: $service_name"
		if ! enable_user_service_without_session "$target_user" "$user_home" "$service_name"; then
			log_error "$service_name konnte nicht aktiviert werden."
			return 1
		fi
		log_warn "User session not active; service installed and enabled, start after next login/boot"
		return 0
	fi

	if ! is_graphical_session_active "$target_user" "$user_uid"; then
		log_warn "graphical-session.target ist nicht aktiv; $service_name startet mit der nächsten grafischen Anmeldung."
		return 0
	fi

	if [ "$start_policy" = "defer-start" ]; then
		log_info "Start von $service_name wird bis nach der Fallback-Deaktivierung verschoben."
		return 0
	fi

	log_info "Starting service: $service_name"
	if ! run_user_systemctl "$target_user" "$user_uid" restart "$service_name"; then
		log_error "$service_name konnte nicht gestartet werden."
		return 1
	fi

	log_info "Prüfe Status von $service_name."
	if run_user_systemctl "$target_user" "$user_uid" is-active --quiet "$service_name"; then
		log_success "$service_name läuft als User-Service."
		return 0
	fi

	run_user_systemctl "$target_user" "$user_uid" status "$service_name" --no-pager || true
	log_error "$service_name ist nicht aktiv."
	return 1
}

install_fallback_service() {
	# Install the legacy browser service but keep it disabled and stopped. It
	# remains available for manual fallback if Cage must be bypassed.
	target_user=$1
	user_home=$2
	user_uid=$3
	service_name=$4

	if ! install_service_file "$target_user" "$user_home" "$service_name"; then
		return 1
	fi

	if ! remove_legacy_default_target_link "$target_user" "$user_home" "$service_name"; then
		return 1
	fi

	if is_user_manager_available "$user_uid"; then
		log_info "Reloading user daemon..."
		if ! run_user_systemctl "$target_user" "$user_uid" daemon-reload; then
			log_error "systemctl --user daemon-reload fehlgeschlagen."
			return 1
		fi

		if run_user_systemctl "$target_user" "$user_uid" is-enabled --quiet "$service_name"; then
			log_info "Disabling fallback service: $service_name"
			if ! run_user_systemctl "$target_user" "$user_uid" disable "$service_name"; then
				log_error "$service_name konnte nicht deaktiviert werden."
				return 1
			fi
		else
			log_info "Fallback service ist nicht aktiviert: $service_name"
		fi

		if run_user_systemctl "$target_user" "$user_uid" is-active --quiet "$service_name"; then
			log_info "Stopping fallback service: $service_name"
			if ! run_user_systemctl "$target_user" "$user_uid" stop "$service_name"; then
				log_error "$service_name konnte nicht gestoppt werden."
				return 1
			fi
		else
			log_info "Fallback service laeuft nicht: $service_name"
		fi
	else
		if ! disable_user_service_without_session "$target_user" "$user_home" "$service_name"; then
			return 1
		fi
		log_warn "User session not active; fallback service installed and disabled for next login/boot"
	fi

	log_success "$service_name installiert und als Fallback deaktiviert."
	return 0
}

start_runtime_service() {
	# Start the productive Cage runtime only when a graphical session exists.
	target_user=$1
	user_uid=$2
	service_name=kiosk-runtime.service

	if ! is_user_manager_available "$user_uid"; then
		log_warn "User session not active; $service_name starts after next login/boot"
		return 0
	fi

	log_info "Reloading user daemon..."
	if ! run_user_systemctl "$target_user" "$user_uid" daemon-reload; then
		log_error "systemctl --user daemon-reload fehlgeschlagen."
		return 1
	fi

	if ! is_graphical_session_active "$target_user" "$user_uid"; then
		log_warn "graphical-session.target ist nicht aktiv; $service_name startet mit der naechsten grafischen Anmeldung."
		return 0
	fi

	log_info "Starting service: $service_name"
	if ! run_user_systemctl "$target_user" "$user_uid" restart "$service_name"; then
		log_error "$service_name konnte nicht gestartet werden."
		return 1
	fi

	log_info "Pruefe Status von $service_name."
	if run_user_systemctl "$target_user" "$user_uid" is-active --quiet "$service_name"; then
		log_success "$service_name laeuft als produktive Appliance Runtime."
		return 0
	fi

	run_user_systemctl "$target_user" "$user_uid" status "$service_name" --no-pager || true
	log_error "$service_name ist nicht aktiv."
	return 1
}

install_user_services() {
	# Install and enable currently active kiosk-client user services.
	target_user=$(detect_target_user)
	user_home=$(get_user_home "$target_user")
	user_uid=$(get_user_uid "$target_user")

	for service_name in $ENABLED_USER_SERVICES; do
		if [ "$service_name" = "kiosk-runtime.service" ]; then
			install_user_service "$target_user" "$user_home" "$user_uid" "$service_name" "defer-start"
		else
			install_user_service "$target_user" "$user_home" "$user_uid" "$service_name"
		fi
	done

	for service_name in $FALLBACK_USER_SERVICES; do
		install_fallback_service "$target_user" "$user_home" "$user_uid" "$service_name"
	done

	start_runtime_service "$target_user" "$user_uid"
}

main() {
	install_user_services
}

if [ "${0##*/}" = "systemd.sh" ]; then
	main "$@"
fi
