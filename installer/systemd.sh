#!/bin/sh
#
# systemd user-service setup module for kiosk-client.
#
# Purpose:
#   Installs and enables kiosk-browser.service as a systemd user service. The
#   browser runs in the graphical user session, while the main installer may run
#   as root.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
SERVICE_NAME="kiosk-browser.service"
SERVICE_SOURCE="$PROJECT_DIR/systemd/user/$SERVICE_NAME"

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

enable_user_service_without_session() {
	# Enable the user service by creating the standard default.target.wants link.
	# This is used when no active user session exists yet and systemctl --user
	# cannot talk to the user manager.
	target_user=$1
	user_home=$2
	user_systemd_dir=$user_home/.config/systemd/user
	wants_dir=$user_systemd_dir/default.target.wants
	service_target=$user_systemd_dir/$SERVICE_NAME
	wants_link=$wants_dir/$SERVICE_NAME

	if ! mkdir -p "$wants_dir"; then
		log_error "User-Service-Aktivierungsverzeichnis konnte nicht erstellt werden: $wants_dir"
		return 1
	fi

	if [ ! -e "$wants_link" ] && ! ln -s "../$SERVICE_NAME" "$wants_link"; then
		log_error "$SERVICE_NAME konnte nicht für default.target aktiviert werden."
		return 1
	fi

	if ! chown -h "$target_user:$target_user" "$wants_dir" "$wants_link" "$service_target"; then
		log_error "Besitzrechte für die User-Service-Aktivierung konnten nicht gesetzt werden."
		return 1
	fi

	return 0
}

install_service_file() {
	# Copy the user service and assign ownership to the target user.
	target_user=$1
	user_home=$2
	user_systemd_dir=$user_home/.config/systemd/user
	service_target=$user_systemd_dir/$SERVICE_NAME

	log_info "Installing user service..."
	if ! mkdir -p "$user_systemd_dir"; then
		log_error "User-systemd-Verzeichnis konnte nicht erstellt werden: $user_systemd_dir"
		return 1
	fi

	if ! cp "$SERVICE_SOURCE" "$service_target"; then
		log_error "$SERVICE_NAME konnte nicht installiert werden."
		return 1
	fi

	if ! chown "$target_user:$target_user" "$user_systemd_dir" "$service_target"; then
		log_error "Besitzrechte für $SERVICE_NAME konnten nicht gesetzt werden."
		return 1
	fi

	if ! chmod 0644 "$service_target"; then
		log_error "Dateirechte für $SERVICE_NAME konnten nicht gesetzt werden."
		return 1
	fi

	log_success "$SERVICE_NAME installiert: $service_target"
	return 0
}

install_browser_service() {
	# Install, enable, restart, and verify the systemd user service.
	if [ ! -r "$SERVICE_SOURCE" ]; then
		log_error "Service-Datei nicht lesbar: $SERVICE_SOURCE"
		return 1
	fi

	target_user=$(detect_target_user)
	user_home=$(get_user_home "$target_user")
	user_uid=$(get_user_uid "$target_user")

	if ! install_service_file "$target_user" "$user_home"; then
		return 1
	fi

	if [ -d "/run/user/$user_uid" ]; then
		log_info "Reloading user daemon..."
		if ! run_user_systemctl "$target_user" "$user_uid" daemon-reload; then
			log_error "systemctl --user daemon-reload fehlgeschlagen."
			return 1
		fi

		log_info "Enabling service..."
		if ! run_user_systemctl "$target_user" "$user_uid" enable "$SERVICE_NAME"; then
			log_error "$SERVICE_NAME konnte nicht aktiviert werden."
			return 1
		fi
	else
		log_info "Enabling service..."
		if ! enable_user_service_without_session "$target_user" "$user_home"; then
			log_error "$SERVICE_NAME konnte nicht aktiviert werden."
			return 1
		fi
		log_warn "User session not active; service installed and enabled, start after next login/boot"
		return 0
	fi

	log_info "Starting service..."
	if ! run_user_systemctl "$target_user" "$user_uid" restart "$SERVICE_NAME"; then
		log_error "$SERVICE_NAME konnte nicht gestartet werden."
		return 1
	fi

	log_info "Prüfe Status von $SERVICE_NAME."
	if run_user_systemctl "$target_user" "$user_uid" is-active --quiet "$SERVICE_NAME"; then
		log_success "$SERVICE_NAME läuft als User-Service."
		return 0
	fi

	run_user_systemctl "$target_user" "$user_uid" status "$SERVICE_NAME" --no-pager || true
	log_error "$SERVICE_NAME ist nicht aktiv."
	return 1
}

main() {
	install_browser_service
}

if [ "${0##*/}" = "systemd.sh" ]; then
	main "$@"
fi
