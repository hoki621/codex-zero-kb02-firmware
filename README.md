# codex-zero-kb02-firmware

[日本語](README_JA.md) · [System setup and controls](https://github.com/hoki621/codex-zero-kb02#readme)

TinyGo input/display firmware for zero-kb02. It sends physical K1–K12 and signed encoder deltas over USB CDC major 2. The joystick uses standard HID mouse; keyboard output, Vial and push inputs are disabled. Herdr-specific actions belong to Host.

## Build and flash

From the parent repository, use the versions in `mise.toml`: **TinyGo 0.40.1 / Go 1.25.13**.

```sh
cd firmware
go test ./...
go vet ./...
tinygo build -o /tmp/zero-kb02.uf2 --target waveshare-rp2040-zero -tags kb02_inputonly --stack-size 8kb --size short .
```

Use `mise exec --` with the commands if the pinned tools are not on PATH. `kb02_inputonly` is required: it excludes upstream HID keyboard/Vial initialization. Building does not flash a device.

For flashing, stop the bridge and serial monitors, retain the [recovery UF2 and procedure](docs/hardware-diagnostics.md), then enter BOOTSEL mode and copy `/tmp/zero-kb02.uf2` to the `RPI-RP2` drive. After reboot the USB serial identifier is `zero-kb02-v2`. If an agent performs the operation, explicitly authorize the target and flash operation.

## Source and behavior

- Matrix scan/debounce: sago35/tinygo-keyboard at `cf173e98f60329b7f7feba941461bb95c065c418`, MIT. The vendored matrix algorithm is unchanged; the adjacent patch isolates input-only code.
- Pin and LED mapping: MIT-licensed sago35/keyboards zero-kb02 at `4b18114b66637c5909229704c0a503bcdeccc057`, marked at the mapping definitions.
- Encoder, SSD1306, WS2812B and drawing/font APIs: pinned public libraries in `go.mod`.
- Local code: major 2 parser/session, bounded queues, joystick calibration and six-slot rendering. No workshop source is copied.

Full notices are retained in [third_party](third_party/README.md) and [recovery](recovery/sago35-keyboards-LICENSE.txt). Preserve the TomThumb font notice when distributing binaries.

The Host must send HELLO and an online STATE before inputs are enabled. Invalid lines do not refresh the 12-second heartbeat. Overflow, offline and reconnect discard queued input; held keys must be released before use. LED brightness is capped at 16/255. Joystick center, dead zones and inversion live in `input.go`; encoder precision/direction and matrix polarity live in `main.go`.

Tests cover protocol, queues, input filtering and in-memory rendering. The [verification record](https://github.com/hoki621/codex-zero-kb02/blob/main/docs/verification.md) separates builds from physical observations. USB CDC writes are buffered by TinyGo and do not acknowledge delivery; display transfers share the scan loop.
