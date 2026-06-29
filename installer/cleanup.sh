#!/bin/sh
#
# Installer cleanup module for kiosk-client.
#
# Purpose:
#   Prepares the future cleanup structure for temporary installer artifacts.
#   This file intentionally contains no productive cleanup commands yet.

set -eu

cleanup_temporary_files() {
	# TODO: Remove temporary files created by future installer steps.
	:
}

cleanup_package_cache() {
	# TODO: Optionally clean package caches when explicitly requested.
	:
}

print_cleanup_summary() {
	# TODO: Report cleanup actions once they exist.
	:
}
