#!/usr/bin/env bash
set -euo pipefail

readonly expected_commit=ed77415774e25e3adf72d192543b14d10438fb1d
readonly expected_sha=a66a99233b4b52de447242eb7a13023e7eb8039727f3966befef12fd1b607389
readonly source_dir=${1:?usage: build-recovery.sh /path/to/sago35-keyboards}
readonly script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
readonly repo_dir=$(dirname -- "$script_dir")
readonly output_dir="$repo_dir/recovery"
readonly output_name=zero-kb02-v0.10.0.uf2
readonly output="$output_dir/$output_name"
readonly temporary_output=$(mktemp "${TMPDIR:-/tmp}/zero-kb02-recovery.XXXXXX.uf2")
readonly actual_commit=$(git -C "$source_dir" rev-parse HEAD)
readonly tinygo_version=$(tinygo version)
readonly go_version=$(go version)
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

mkdir -p "$output_dir"
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
test "$actual_sha" = "$expected_sha" || {
  echo "recovery UF2 is not byte-reproducible: expected $expected_sha, got $actual_sha" >&2
  exit 1
}
cp "$temporary_output" "$output"
cp "$source_dir/LICENSE.txt" "$output_dir/sago35-keyboards-LICENSE.txt"
(cd "$output_dir" && shasum -a 256 "$output_name" > "$output_name.sha256")
