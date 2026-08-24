#!/usr/bin/env bash
set -euo pipefail

readonly script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
export CAPTURE_SCRIPT="$script_dir/capture-joystick-diagnostic.exp"

run_mock() {
  MOCK_OUTPUT=$1 /usr/bin/expect -c '
    rename spawn real_spawn
    rename log_file real_log_file
    proc log_file args {}
    proc spawn args {
      uplevel #0 [list real_spawn /bin/sh -c {
        printf "%s" "$MOCK_OUTPUT"
        sleep 2
      }]
    }
    source $env(CAPTURE_SCRIPT)
  '
}

valid='JOYSTICK_DIAGNOSTIC v2 DTR 1
SUMMARY center 100 1 2 1 3 4 3
SUMMARY up 40 1 2 1 3 4 3
SUMMARY down 40 1 2 1 3 4 3
SUMMARY left 40 1 2 1 3 4 3
SUMMARY right 40 1 2 1 3 4 3
DONE
'
run_mock "$valid"

set +e
run_mock 'JOYSTICK_DIAGNOSTIC v2 DTR 1
SUMMARY center 100
'
readonly truncated_status=$?
set -e
test "$truncated_status" -eq 2
