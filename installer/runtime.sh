#!/bin/sh
#
# Appliance runtime setup for kiosk-client.
#
# Purpose:
#   Installs and enables the Appliance Edition systemd user services. This
#   runtime does not require a graphical login manager or full graphical environment.

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
PROJECT_DIR=$(CDPATH= cd "$SCRIPT_DIR/.." && pwd)
APPLIANCE_USER_SERVICES="
kiosk-agent.service
kiosk-appliance.service
"
AGENT_BINARY="$PROJECT_DIR/kiosk-agent"
AGENT_SOURCE="./agent/cmd/kiosk-agent"
AGENT_INSTALL_MODE=release

# shellcheck disable=SC1091
. "$SCRIPT_DIR/install-common.sh"

detect_kiosk_user() {
	# Determine the user account that owns the appliance user service.
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

get_user_home() {
	# Resolve the target user's home directory.
	kiosk_user=$1
	home_dir=$(getent passwd "$kiosk_user" | awk -F ':' '{ print $6 }')

	if [ "$home_dir" != "" ]; then
		printf '%s\n' "$home_dir"
		return 0
	fi

	log_error "Home-Verzeichnis fuer Benutzer '$kiosk_user' konnte nicht ermittelt werden."
	return 1
}

get_user_uid() {
	# Resolve the target user's numeric uid.
	kiosk_user=$1

	if user_uid=$(id -u "$kiosk_user" 2>/dev/null); then
		printf '%s\n' "$user_uid"
		return 0
	fi

	log_error "UID fuer Benutzer '$kiosk_user' konnte nicht ermittelt werden."
	return 1
}

run_user_systemctl() {
	# Run systemctl --user as the kiosk user when a user manager is reachable.
	kiosk_user=$1
	user_uid=$2
	shift 2

	sudo -u "$kiosk_user" XDG_RUNTIME_DIR="/run/user/$user_uid" systemctl --user "$@"
}

ensure_agent_binary() {
	# Ensure kiosk-agent exists before installing or enabling its service.
	kiosk_user=$1

	if command -v go >/dev/null 2>&1; then
		AGENT_INSTALL_MODE=development
		log_info "Development Mode: Baue kiosk-agent aus vorhandenen Quellen."
		if ! (cd "$PROJECT_DIR" && go build -o "$AGENT_BINARY" "$AGENT_SOURCE"); then
			log_error "kiosk-agent konnte nicht gebaut werden."
			return 1
		fi
	else
		log_info "Release Mode: Go ist nicht installiert; verwende vorhandenes kiosk-agent Binary."
		if [ -x "$AGENT_BINARY" ]; then
			log_success "kiosk-agent Binary vorhanden: $AGENT_BINARY"
			return 0
		fi

		if [ -e "$AGENT_BINARY" ]; then
			log_error "kiosk-agent existiert, ist aber nicht ausfuehrbar: $AGENT_BINARY"
			return 1
		fi

		log_error "kiosk-agent fehlt: $AGENT_BINARY"
		log_error "Go ist nicht installiert; bitte ein Release mit vorhandenem kiosk-agent Binary verwenden."
		return 1
	fi

	if [ ! -x "$AGENT_BINARY" ] && [ -e "$AGENT_BINARY" ]; then
		log_error "kiosk-agent existiert, ist aber nicht ausfuehrbar: $AGENT_BINARY"
		return 1
	fi

	if ! chmod 0755 "$AGENT_BINARY"; then
		log_error "Dateirechte fuer kiosk-agent konnten nicht gesetzt werden."
		return 1
	fi

	if ! chown "$kiosk_user:$kiosk_user" "$AGENT_BINARY"; then
		log_error "Besitzrechte fuer kiosk-agent konnten nicht gesetzt werden."
		return 1
	fi

	log_success "kiosk-agent gebaut: $AGENT_BINARY"
}

install_service_file() {
	# Copy one Appliance user service into the kiosk user's systemd directory.
	kiosk_user=$1
	user_home=$2
	service_name=$3
	service_source=$PROJECT_DIR/systemd/user/$service_name
	user_systemd_dir=$user_home/.config/systemd/user
	service_target=$user_systemd_dir/$service_name

	if [ ! -r "$service_source" ]; then
		log_error "Service-Datei nicht lesbar: $service_source"
		return 1
	fi

	if ! mkdir -p "$user_systemd_dir"; then
		log_error "User-systemd-Verzeichnis konnte nicht erstellt werden: $user_systemd_dir"
		return 1
	fi

	if ! cp "$service_source" "$service_target"; then
		log_error "$service_name konnte nicht installiert werden."
		return 1
	fi

	if ! chown "$kiosk_user:$kiosk_user" "$user_systemd_dir" "$service_target"; then
		log_error "Besitzrechte fuer $service_name konnten nicht gesetzt werden."
		return 1
	fi

	if ! chmod 0644 "$service_target"; then
		log_error "Dateirechte fuer $service_name konnten nicht gesetzt werden."
		return 1
	fi

	log_success "$service_name installiert: $service_target"
}

enable_service_without_session() {
	# Enable the user service by creating the default.target.wants symlink.
	kiosk_user=$1
	user_home=$2
	service_name=$3
	user_systemd_dir=$user_home/.config/systemd/user
	wants_dir=$user_systemd_dir/default.target.wants
	wants_link=$wants_dir/$service_name

	if ! mkdir -p "$wants_dir"; then
		log_error "Aktivierungsverzeichnis konnte nicht erstellt werden: $wants_dir"
		return 1
	fi

	if [ ! -e "$wants_link" ]; then
		if ! ln -s "../$service_name" "$wants_link"; then
			log_error "$service_name konnte nicht aktiviert werden."
			return 1
		fi
	fi

	if ! chown -h "$kiosk_user:$kiosk_user" "$wants_dir" "$wants_link"; then
		log_error "Besitzrechte fuer die Service-Aktivierung konnten nicht gesetzt werden."
		return 1
	fi

	log_success "$service_name fuer default.target aktiviert."
}

install_appliance_runtime() {
	# Install and enable the appliance user runtime services.
	if ! require_root; then
		return 1
	fi

	kiosk_user=$(detect_kiosk_user)
	user_home=$(get_user_home "$kiosk_user")
	user_uid=$(get_user_uid "$kiosk_user")

	ensure_agent_binary "$kiosk_user"

	for service_name in $APPLIANCE_USER_SERVICES; do
		install_service_file "$kiosk_user" "$user_home" "$service_name"
	done

	if [ -d "/run/user/$user_uid" ]; then
		log_info "Reloading user daemon..."
		run_user_systemctl "$kiosk_user" "$user_uid" daemon-reload
		if [ "$AGENT_INSTALL_MODE" = "development" ]; then
			log_success "Development build completed."
		fi
		for service_name in $APPLIANCE_USER_SERVICES; do
			log_info "Enabling service: $service_name"
			run_user_systemctl "$kiosk_user" "$user_uid" enable "$service_name"
		done
	else
		for service_name in $APPLIANCE_USER_SERVICES; do
			enable_service_without_session "$kiosk_user" "$user_home" "$service_name"
		done
		if [ "$AGENT_INSTALL_MODE" = "development" ]; then
			log_success "Development build completed."
		fi
		log_warn "User session not active; appliance services start after tty1 autologin."
	fi

	log_success "Appliance Runtime installiert."
}

main() {
	install_appliance_runtime
}

if [ "${0##*/}" = "runtime.sh" ]; then
	main "$@"
fi
