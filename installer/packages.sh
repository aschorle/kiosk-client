#!/bin/sh
#
# Package lists for kiosk-client.
#
# Purpose:
#   Defines package groups for later installer phases. This file contains only
#   package list variables and intentionally runs no apt or installation commands.

set -eu

# Minimal Appliance Edition packages.
#
# Keep this list limited to packages required by the runtime path:
# systemd user service -> dbus-run-session -> Cage -> Chromium.
APPLIANCE_PACKAGES="
ca-certificates
chromium
cage
dbus
dbus-user-session
"

# Deprecated package groups are intentionally empty.
COMMON_PACKAGES=""
KIOSK_PACKAGES="$APPLIANCE_PACKAGES"

# Optional development and diagnostics packages. Not used by Appliance Edition.
DEV_PACKAGES=""

# Radxa-specific packages are intentionally undefined in this phase.
RADXA_PACKAGES=""

# Raspberry Pi specific packages are intentionally undefined in this phase.
RPI_PACKAGES=""
