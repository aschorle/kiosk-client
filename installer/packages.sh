#!/bin/sh
#
# Package lists for kiosk-client.
#
# Purpose:
#   Defines package groups for later installer phases. This file contains only
#   package list variables and intentionally runs no apt or installation commands.

set -eu

# Common base packages required by both supported platforms.
COMMON_PACKAGES="
git
curl
wget
vim
openssh-server
ca-certificates
network-manager
"

# Kiosk runtime packages for later browser/display phases.
KIOSK_PACKAGES="
chromium
cage
"

# Optional development and diagnostics packages.
DEV_PACKAGES="
htop
tree
"

# Radxa-specific packages are intentionally undefined in this phase.
RADXA_PACKAGES=""

# Raspberry Pi specific packages are intentionally undefined in this phase.
RPI_PACKAGES=""
