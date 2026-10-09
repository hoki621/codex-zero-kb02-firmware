# codex-zero-kb02-firmware

[日本語](README_JA.md) · [Full setup and controls](https://github.com/hoki621/codex-zero-kb02/blob/main/README_EN.md)

TinyGo firmware for the zero-kb02 Codex controller. It sends key and encoder input to the Mac, displays conversation states on the OLED and LEDs, and uses the joystick to move the mouse pointer. Use it with the [Host application](https://github.com/hoki621/codex-zero-kb02-host).

## Build

Clone the parent repository with its submodules and run `mise install` as described in the [setup guide](https://github.com/hoki621/codex-zero-kb02/blob/main/README_EN.md#setup). The build uses **TinyGo 0.40.1 / Go 1.25.13**. Run these commands from the parent repository root:

```sh
mise exec -- sh -c 'cd firmware && go test ./... && go vet ./...'
mise exec -- sh -c 'cd firmware && tinygo build -o /tmp/zero-kb02.uf2 --target waveshare-rp2040-zero -tags kb02_inputonly --stack-size 8kb --size short .'
```

The output is `/tmp/zero-kb02.uf2`. The `kb02_inputonly` build tag is required: it excludes HID keyboard and Vial initialization so the keys are handled by Host. Building does not flash the device.

## Flash and connect

1. Stop the Host bridge and serial monitors. Keep the [recovery firmware and instructions](docs/hardware-diagnostics.md) available.
2. Put the device in BOOTSEL mode. When the `RPI-RP2` drive appears, copy `/tmp/zero-kb02.uf2` to it.
3. After reboot, the USB serial identifier is `zero-kb02-v2`. Follow the [startup guide](https://github.com/hoki621/codex-zero-kb02/blob/main/README_EN.md#run) to connect Host using the device's actual serial port.

Keys become active when Host connects and sends an online state. After 12 seconds without a valid heartbeat, the display goes offline and the LEDs turn off. Release held keys after reconnecting.

## Development

Firmware uses [USB CDC protocol major 2](https://github.com/hoki621/codex-zero-kb02/blob/main/PROTOCOL.md); major 1 is incompatible. The joystick uses standard HID mouse output. HID keyboard output, Vial and push inputs are disabled.

Joystick calibration, dead zones and direction settings are in `input.go`. Encoder settings and matrix polarity are in `main.go`. LED brightness is capped at 16/255.

Tests cover the protocol, input queues, filtering and in-memory rendering. They do not replace hardware checks; see the [verification record](https://github.com/hoki621/codex-zero-kb02/blob/main/docs/verification.md). Display transfers share the input scan loop, and USB CDC writes do not acknowledge delivery.

## Libraries and licenses

- Matrix scanning and debounce: MIT-licensed [sago35/tinygo-keyboard](https://github.com/sago35/tinygo-keyboard), vendored with an input-only build patch.
- Pin and LED mapping: MIT-licensed [sago35/keyboards](https://github.com/sago35/keyboards), with source comments at the mapping definitions.
- Encoder, display, LED and font support: public libraries pinned in `go.mod`.

The device protocol, joystick calibration and six-slot display mapping are local to this project. Source revisions and complete notices are in [third_party](third_party/README.md) and [recovery](recovery/sago35-keyboards-LICENSE.txt). Retain the upstream licenses and TomThumb font notice when distributing binaries.
