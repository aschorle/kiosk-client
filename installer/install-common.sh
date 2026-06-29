#!/bin/sh
#
# Shared installer helpers for kiosk-client.
#
# Purpose:
#   Provides common functions used by all platform-specific installer scripts.
#   This file intentionally contains no productive installation commands yet.

set -eu

log_info() {
	# TODO: Standardize informational logging.
	:
}

require_root() {
	# TODO: Validate installer privileges before productive actions are added.
	:
}

detect_debian_version() {
	# TODO: Verify Debian 12 Bookworm as the supported base system.
	:
}

load_client_config() {
	# TODO: Read the future kiosk-client configuration file.
	:
}
