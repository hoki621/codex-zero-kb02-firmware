# codex-zero-kb02-firmware

TinyGo firmware for the unofficial
[`codex-zero-kb02`](https://github.com/hoki621/codex-zero-kb02) project.

Implementation work is tracked in the parent repository. This repository owns
zero-kb02 input scanning, RGB/OLED rendering, and the bounded USB CDC protocol.
It contains no Herdr-specific logic.

## Firmware v1

Build the RP2040 UF2 with the pinned TinyGo toolchain:

```sh
tinygo build -o /absolute/output.uf2 --target waveshare-rp2040-zero --stack-size 8kb --size short .
```

The USB CDC implementation follows the parent `PROTOCOL.md`: it requires the
v1 handshake and an online `STATE` before emitting input, bounds each line to
128 bytes, and enters offline display state after 12 seconds without a valid
host message.

The six protocol agent keys are K2, K3, K5, K6, K7, and K8, mapped to slots
0 through 5. K1, K4, and K9-K12 are reserved and emit no v1 event. GP0
joystick press is scanned and debounced but also emits no event because v1 has
no joystick-press message. Encoder press alone uses `ENC ... DOWN|UP`.

Slot RGB is shown only on the six agent-key LEDs at a maximum channel value of
16/255. OLED I2C runs at 400kHz and the framebuffer is transmitted only when
the six-slot panel state changes. Joystick calibration is defined once in
`input.go`.
