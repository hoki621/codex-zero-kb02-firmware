#!/usr/bin/env bash
set -euo pipefail

readonly expected_commit=ed77415774e25e3adf72d192543b14d10438fb1d
readonly source_dir=${1:?usage: build-recovery.sh /path/to/sago35-keyboards /absolute/output.uf2}
readonly output=${2:?usage: build-recovery.sh /path/to/sago35-keyboards /absolute/output.uf2}
readonly script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
readonly repo_dir=$(dirname -- "$script_dir")
readonly committed_output="$repo_dir/recovery/zero-kb02-v0.10.0.uf2"
readonly license_output="$output.LICENSE.txt"
readonly actual_commit=$(git -C "$source_dir" rev-parse HEAD)
readonly tinygo_version=$(tinygo version)
readonly go_version=$(go version)

[[ "$output" = /*.uf2 ]] || {
  echo "output must be an absolute .uf2 path: $output" >&2
  exit 1
}
test "$output" != "$committed_output" || {
  echo "refusing to overwrite committed recovery artifact: $output" >&2
  exit 1
}
test ! -e "$output" && test ! -e "$output.sha256" && test ! -e "$license_output" || {
  echo "refusing to overwrite existing output: $output" >&2
  exit 1
}

readonly output_dir=$(dirname -- "$output")
mkdir -p "$output_dir"
readonly temporary_output=$(mktemp "$output_dir/.zero-kb02-recovery.XXXXXX.uf2")
trap 'rm -f "$temporary_output"' EXIT

test "$actual_commit" = "$expected_commit" || {
  echo "expected source commit $expected_commit, got $actual_commit" >&2
  exit 1
}
[[ "$tinygo_version" == 'tinygo version 0.40.1 '* ]] || {
  echo "expected TinyGo 0.40.1, got: $tinygo_version" >&2
  exit 1
}
[[ "$go_version" == 'go version go1.25.13 '* ]] || {
  echo "expected Go 1.25.13, got: $go_version" >&2
  exit 1
}

(
  cd "$source_dir"
  GO111MODULE=on tinygo build -o "$temporary_output" \
    --target waveshare-rp2040-zero \
    --size short \
    --stack-size 8kb \
    --tags zero_kb02 \
    ./zero-kb02/firmware/
)
actual_sha=$(shasum -a 256 "$temporary_output" | awk '{print $1}')
mv "$temporary_output" "$output"
printf '%s  %s\n' "$actual_sha" "$(basename -- "$output")" > "$output.sha256"
cp "$source_dir/LICENSE.txt" "$license_output"
