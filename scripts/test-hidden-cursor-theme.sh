#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd "$(dirname "$0")" && pwd)
START_SCRIPT=$SCRIPT_DIR/start-cage.sh
TEST_ROOT=$(mktemp -d)
trap 'rm -rf "$TEST_ROOT"' EXIT HUP INT TERM

mkdir -p "$TEST_ROOT/bin" "$TEST_ROOT/runtime"
printf '%s\n' '#!/bin/sh' 'exit 0' > "$TEST_ROOT/bin/cage-test"
chmod 0755 "$TEST_ROOT/bin/cage-test"

# The production runtime is Linux. Git Bash on Windows may emulate `ln -s`
# as a regular file when symlink privileges are unavailable, so report the
# test as skipped there instead of asserting against emulated filesystem state.
printf '%s\n' probe > "$TEST_ROOT/probe-target"
ln -s probe-target "$TEST_ROOT/probe-link"
if [ ! -L "$TEST_ROOT/probe-link" ]; then
	printf '%s\n' 'hidden cursor theme creation and repair: skipped (symlinks unavailable)'
	exit 0
fi

run_theme_creation() {
	XDG_RUNTIME_DIR="$TEST_ROOT/runtime" \
	CAGE_BIN=cage-test \
	PATH="$TEST_ROOT/bin:$PATH" \
		"$START_SCRIPT" >/dev/null
}

cursor_dir=$TEST_ROOT/runtime/kiosk-client-cursors/kiosk-hidden/cursors
cursor_file=$cursor_dir/left_ptr

run_theme_creation
[ ! -L "$cursor_file" ]
[ -f "$cursor_file" ]
[ "$(od -An -tx1 -N4 "$cursor_file" | tr -d ' \n')" = "58637572" ]

for cursor_name in default arrow right_ptr pointer hand1 text; do
	[ -L "$cursor_dir/$cursor_name" ]
	[ "$(readlink "$cursor_dir/$cursor_name")" = "left_ptr" ]
done

ln -sf left_ptr "$cursor_file"
[ -L "$cursor_file" ]
run_theme_creation
[ ! -L "$cursor_file" ]
[ -f "$cursor_file" ]
[ "$(od -An -tx1 -N4 "$cursor_file" | tr -d ' \n')" = "58637572" ]

run_theme_creation
[ ! -L "$cursor_file" ]
[ "$(readlink "$cursor_dir/arrow")" = "left_ptr" ]

printf '%s\n' 'hidden cursor theme creation and repair: ok'
