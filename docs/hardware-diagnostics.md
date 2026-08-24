# zero-kb02 hardware diagnostics

Issue #4 の診断記録。`PASS` はこの文書に証跡がある確認だけを示す。
実機操作を伴う項目は、明示許可を得るまで `NOT TESTED` のままにする。

## 接続情報

2026-08-24 に macOS の IORegistry と device node を read-only で確認した。
serial port は開いていない。

| Field | Value |
| --- | --- |
| Product | `zero-kb02-0.1.0` |
| Manufacturer | `Waveshare` |
| VID:PID | `2E8A:0003` |
| Serial | `vial:f64c2b3c` |
| Callout path | `/dev/cu.usbmodemvial_f64c2b3c1` |
| TTY path | `/dev/tty.usbmodemvial_f64c2b3c1` |

再確認は device node を開かない次のコマンドで行う。

```sh
ioreg -p IOUSB -l -w 0
ls -l /dev/cu.* /dev/tty.*
```

## GPIO 割り当て

配線は sago35/keyboards の公開 pinout と回路説明を照合した静的情報であり、
この個体では未確認。

| Part | GPIO | Mode / note |
| --- | --- | --- |
| Joystick press | GP0 | input, pull-up |
| 12 x RGB LED data | GP1 | output, serial RGB chain |
| Encoder press | GP2 | input, pull-up |
| Encoder A | GP3 | input, pull-up |
| Encoder B | GP4 | input, pull-up |
| Matrix COL1..COL4 | GP5..GP8 | output |
| Matrix ROW1..ROW3 | GP9..GP11 | input, pull-down |
| OLED SDA | GP12 | I2C0 SDA |
| OLED SCL | GP13 | I2C0 SCL |
| Joystick Y | GP28 / ADC2 | analog input |
| Joystick X | GP29 / ADC3 | analog input |

Sources:

- <https://github.com/sago35/keyboards/tree/ed77415774e25e3adf72d192543b14d10438fb1d/zero-kb02>
- <https://github.com/tinygo-keeb/workshop/blob/main/README_EN.md>

The workshop is reference-only. No workshop source code is copied here.

## Recovery UF2 build

| Field | Value |
| --- | --- |
| Source | <https://github.com/sago35/keyboards> |
| License | MIT (`LICENSE.txt` in the source repository) |
| Tag | `v0.10.0` |
| Commit | `ed77415774e25e3adf72d192543b14d10438fb1d` |
| Source path | `zero-kb02/firmware/` |
| TinyGo | `0.40.1 darwin/arm64`, LLVM `20.1.1` |
| Go | `go1.25.13 darwin/arm64` |
| Output | `recovery/zero-kb02-v0.10.0.uf2` (417792 bytes) |
| SHA-256 | `a66a99233b4b52de447242eb7a13023e7eb8039727f3966befef12fd1b607389` |
| Build result | PASS (`code 208728`, `data 0`, `bss 46336`) |

Build from a clean checkout and reject any unexpected commit before compiling.

```sh
git clone --branch v0.10.0 --depth 1 https://github.com/sago35/keyboards.git
cd keyboards
test "$(git rev-parse HEAD)" = ed77415774e25e3adf72d192543b14d10438fb1d
mkdir -p out
GO111MODULE=on tinygo build -o ./out/zero-kb02.uf2 --target waveshare-rp2040-zero --size short --stack-size 8kb --tags zero_kb02 ./zero-kb02/firmware/
shasum -a 256 ./out/zero-kb02.uf2
```

From this repository, `scripts/build-recovery.sh /path/to/keyboards` performs
the commit and tool-version checks and builds to a temporary file. It updates
the UF2 plus `.sha256` under `recovery/` only when the result matches the
recorded SHA-256; a mismatch fails without overwriting the recovery artifact.
It also copies the upstream MIT license beside the artifact. It does not clone,
flash, or open a device.

The checkout, tool versions, build result, output size, and SHA-256 were
verified on 2026-08-24. No flash or device operation was performed.

## Joystick diagnostic firmware

This repository includes a finite, non-HID diagnostic at
`diagnostics/joystick/`. It samples GP29 (X) and GP28 (Y) and has no keyboard,
mouse, LED, OLED, or command input.

| Field | Value |
| --- | --- |
| Build command | `GO111MODULE=off tinygo build -o recovery/zero-kb02-joystick-diagnostic.uf2 --target waveshare-rp2040-zero --size short ./diagnostics/joystick/` |
| USB product | `zero-kb02-joystick-diag` |
| USB serial | `zero-kb02-diag` |
| Output | `recovery/zero-kb02-joystick-diagnostic.uf2` (24064 bytes) |
| SHA-256 | `3e9ca828b2664c67177622239106f8847fe37cbb704155a0e0c5af807a55c06c` |
| Build result | PASS (`code 11748`, `data 108`, `bss 5224`, `flash 11856`, `ram 5332`) |

The firmware waits without emitting samples until the exact serial monitor
asserts USB CDC DTR. It then repeats its `JOYSTICK_DIAGNOSTIC v2` banner three
times before capturing 100 center samples and 40 samples each for up, down,
left, and right. Every phase has a three-second prepare countdown. It prints
raw samples and a min/max/mean summary, prints `DONE` after about 29 seconds,
then produces no more output. The dedicated capture script opens only
`/dev/cu.usbmodemzero_kb02_diag1`; one 60-second timer covers the complete
session. It succeeds only after the banner, ordered summaries with counts
100/40/40/40/40, and `DONE`; timeout returns 124 and malformed or incomplete
output returns 2.

The first diagnostic attempt on 2026-08-24 flashed SHA-256
`1205f9084cbff078eeb75270040b9be0d57a2ec22d30e2736cda56d3bbb80f70`
and verified the expected diagnostic USB identity and exact port. The monitor
connected to that port, but received zero diagnostic lines before its
60-second timeout (exit 124). This attempt is `FAIL`; no joystick measurements
were inferred from it. The DTR handshake above was added after that failure and
was exercised successfully in the retry below.

The retry flashed SHA-256
`3e9ca828b2664c67177622239106f8847fe37cbb704155a0e0c5af807a55c06c`,
verified the diagnostic identity and exact port, and completed with `DONE` and
monitor exit 0. The user moved the joystick in the known physical order center,
up, down, left, right. Command-start latency shifted the physical operations by
about one printed phase, so the `SUMMARY` phase names are not physical-direction
labels. Some sample lines also lost digits. The measurements below use only
long, stable clusters and the known operation order; isolated malformed values
and the printed aggregate summaries are excluded.

## Key and LED order

The matrix wiring establishes the key numbers below. The fixed upstream source
maps keys to zero-based LED indices as shown, but physical one-at-a-time testing
is still required before marking the mapping confirmed.

| Matrix | COL1 | COL2 | COL3 | COL4 |
| --- | --- | --- | --- | --- |
| ROW1 | key 1 | key 2 | key 3 | key 4 |
| ROW2 | key 5 | key 6 | key 7 | key 8 |
| ROW3 | key 9 | key 10 | key 11 | key 12 |

| Key | Source LED index | Physical LED | Key result | LED result |
| --- | --- | --- | --- | --- |
| 1 | 0 | unknown | PASS: `a` | NOT REPORTED |
| 2 | 3 | unknown | PASS: `b` | NOT REPORTED |
| 3 | 6 | unknown | PASS: `c` | NOT REPORTED |
| 4 | 9 | unknown | PASS: `d` | NOT REPORTED |
| 5 | 1 | unknown | PASS: `e` | NOT REPORTED |
| 6 | 4 | unknown | PASS: `f` | NOT REPORTED |
| 7 | 7 | unknown | PASS: `g` | NOT REPORTED |
| 8 | 10 | unknown | PASS: `h` | NOT REPORTED |
| 9 | 2 | unknown | PASS: layer modifier | NOT REPORTED |
| 10 | 5 | unknown | PASS: layer modifier | NOT REPORTED |
| 11 | 8 | unknown | PASS: left click | NOT REPORTED |
| 12 | 11 | unknown | PASS: right click | NOT REPORTED |

## Pass/fail record

| Component | Check | Result | Evidence / remaining action |
| --- | --- | --- | --- |
| USB device | enumerate without opening a port | PASS | IORegistry reports `2E8A:0003`, serial and product above |
| Recovery | BOOTSEL + RST, UF2 copy, reboot | PASS | `/dev/disk4s1` mounted at `/Volumes/RPI-RP2`; runtime device re-enumerated after UF2 transfer |
| Keys 1-12 | each expected action occurs | PASS | K1-K8 produced `abcdefgh`; user confirmed K9-K12 matched the current firmware's expected actions |
| LEDs 1-12 | one-at-a-time RGB test | NOT TESTED | flash/device-control permission required |
| OLED | full-frame and orientation test | NOT TESTED | flash/device-control permission required |
| Encoder press | press/release events | PASS | user confirmed expected current-firmware behavior |
| Encoder rotation | CW/CCW direction and detents | PARTIAL | user reported no problem; CW/CCW effects not recorded |
| Joystick press | press/release events | NOT TESTED | diagnostic firmware and serial-open permission required |
| Joystick X/Y | center and four directions | PASS | stable clusters establish right=+X, left=-X, up=+Y, down=-Y; center envelope recorded below |
| Diagnostic serial capture | center plus four-direction numeric samples | PASS WITH CAVEAT | retry reached `DONE`/exit 0; physical directions inferred from stable clusters and known order because labels lagged by about one phase |

## Encoder direction candidate

The fixed upstream source uses encoder A=GP3 and B=GP4 by default. Its
`invert_rotary_pins` build tag swaps those pins. Which configuration produces
clockwise and counter-clockwise events on this unit is `NOT TESTED`.

## Joystick calibration candidates

The fixed upstream source uses raw range `0x3000..0xC800`, whose midpoint is
`0x7C00` (31744). The retry and a prior untouched-center run produced these
stable physical clusters:

| Physical position | Approximate raw X | Approximate raw Y |
| --- | ---: | ---: |
| Center, retry release | 31200 | 32560 |
| Center, prior untouched run | 31964 | 32455 |
| Up | 26700 | 57900 |
| Down | 32900 | 8250 |
| Left | 8000 | 31400 |
| Right | 55900 | 34600 |

Across the two runs, the observed released-center envelope was approximately
X=31072..32096 and Y=32320..32672. Its midpoint gives the proposed calibration
center X=31584, Y=32496. Raw right is +X and raw up is +Y; therefore the current
firmware's X non-inverted and Y inverted output mapping is correct for HID
right/down-positive coordinates.

| Setting | Candidate | Status |
| --- | --- | --- |
| X center | 31584 | PROPOSED FROM MEASURED ENVELOPE |
| Y center | 32496 | PROPOSED FROM MEASURED ENVELOPE |
| X inversion | false | CONFIRMED: right increases raw X |
| Y inversion | true | CONFIRMED: up increases raw Y and must map negative |
| Enter threshold | center +/- 8192 | PROPOSED; separates every stable direction cluster |
| Release threshold | center +/- 4096 | PROPOSED; contains the full observed center envelope |

The proposed release band is much wider than the observed run-to-run center
shift, and the enter band remains well below the nearest directional
displacement. Chatter exactly at either threshold was not exercised, so these
remain calibration candidates rather than a product-firmware change.

## BOOTSEL + RST recovery

This procedure was performed once with explicit permission on 2026-08-24.

1. Verify the recovery UF2 SHA-256 against the recorded build output.
2. Hold BOOTSEL, press and release RST, then release BOOTSEL.
3. Confirm that the expected RP2040 mass-storage volume appears at the explicit
   path `/Volumes/RPI-RP2`. Do not select a volume by wildcard.
4. Copy the verified UF2 to that explicit volume path.
5. After automatic reboot, confirm `2E8A:0003` and the expected serial in
   IORegistry. Do not open the serial port as part of this check.

Observed result:

- Bootloader: `RP2 Boot`, VID:PID `2E8A:0003`, serial `E0C9125B0D9B`.
- Mount: `/dev/disk4s1` on `/Volumes/RPI-RP2` (`msdos`).
- Artifact SHA-256 immediately before copy:
  `a66a99233b4b52de447242eb7a13023e7eb8039727f3966befef12fd1b607389`.
- The `cp` process returned 1 while copying an extended attribute after the UF2
  data transfer. The boot volume automatically unmounted and the runtime
  firmware re-enumerated, demonstrating that the RP2040 accepted and booted the
  UF2. No second copy was attempted.
- Runtime after reboot: product `zero-kb02-0.1.0`, VID:PID `2E8A:0003`, serial
  `vial:f64c2b3c`, callout path `/dev/cu.usbmodemvial_f64c2b3c1`.

The recovery procedure is `PASS`. Serial, HID, and Herdr input were not used.

## TinyGo flash rebuild reproducibility

Two source-flash restores rebuilt the same recovery source with TinyGo 0.40.1
and restored the expected runtime product, serial, and path. Their saved
temporary 417792-byte UF2 hashes were
`17bcf5da8a0b417984dab1759bdfdb4b59a02fff9f2d86061e633ea4d1de48cf`
and `5910e41ed7ac0bd12ba698821376714d6d28aaae952c3ab917fe39b0896a84a2`,
neither equal to the stored artifact's `a66a9923...7389`.
Runtime restore therefore passed twice, but byte-for-byte reproduction by
`tinygo flash -o` failed.

Two consecutive read-only builds with explicit `GO111MODULE=on` initially
reproduced the stored `a66a9923...7389` artifact. Repeating the exact command
three times later, with the same output path, produced three different hashes
(`1ae7befb...3015`, `9ec9a9e3...7aac`, and `6fe0b505...50df`). The byte order is
therefore not deterministic. `GO111MODULE=off` did not produce an alternate
hash; it failed because the pinned module dependencies were unavailable in
GOPATH. Comparing the 417792-byte UF2 files found reordered embedded PNG
payloads (for example, the 640x360 and 20x25 IHDR records trade their leading
positions) plus their corresponding references, while size and TinyGo size
metrics stayed equal. Thus module mode omission was not the hash cause, and a
fresh source build or flash must not be claimed byte-identical to the stored
UF2.

Recovery build and source-flash commands use explicit `GO111MODULE=on` and the
same target, size, stack, and tag flags. The temporary flash hash is recorded,
but recovery success is based on the fixed commit/tool checks and the final
read-only runtime identity. Exact recovery to the stored artifact hash remains
the BOOTSEL mass-storage copy procedure above.
