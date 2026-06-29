#!/bin/sh
#
# kiosk-client installer entry point.
#
# Purpose:
#   Coordinates the future installer workflow and selects the target platform.
#   This file intentionally contains no productive installation commands yet.

set -eu

main() {
	# TODO: Parse command line options for target platform and dry-run mode.
	# TODO: Load common installer helpers from install-common.sh.
	# TODO: Dispatch to install-radxa.sh or install-rpi.sh.
	# TODO: Call verification and cleanup steps once productive installers exist.
	:
}

main "$@"
