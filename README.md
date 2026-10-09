# codex-zero-kb02-firmware

TinyGo firmware for the unofficial
[`codex-zero-kb02`](https://github.com/hoki621/codex-zero-kb02) project.

Implementation work is tracked in the parent repository. This repository owns
zero-kb02 input scanning, RGB/OLED rendering, and the bounded USB CDC protocol.
It contains no Herdr-specific logic.

## Firmware major 2

This local implementation follows the parent `PROTOCOL.md` major 2. It requires
`HELLO HOST 2` and a valid online STATE before sending physical `KEY` 1–12
DOWN/UP and signed `ENC` deltas. K11 is reported as a physical key; its Host
binding remains unassigned. Escape, popup, approval and new-chat meanings belong
to Host. Neither push input nor joystick CDC messages are sent. Joystick uses
standard HID mouse only, at three pixels per 10ms outside the calibrated dead zone.

Input queue capacity is 32 lines, one line drained per scan. Overflow discards
queued input and requires another handshake. Startup, reconnect, online/offline
and generation changes discard pending input and suppress held keys until release,
including presses still being debounced. Invalid lines do not refresh heartbeat;
timeout at 12 seconds returns offline and requires HELLO again.

Build with Go 1.25.13 and TinyGo 0.40.1:

```sh
go test ./...
go vet ./...
tinygo build -o /absolute/output.uf2 --target waveshare-rp2040-zero -tags kb02_inputonly --stack-size 8kb --size short .
```

## Library composition

Matrix scan/debounce uses sago35/tinygo-keyboard at fixed commit
`cf173e98f60329b7f7feba941461bb95c065c418` with the existing input-only
patch in `third_party/`. The original matrix/options sources are unchanged.
The zero-kb02 pin and LED order comes from the MIT-licensed
`sago35/keyboards` hardware/firmware reference. The workshop examples are
API references only; they have no confirmed reusable license, so their code
is not copied.
`MatrixKeyboard.Get()` provides the physical K1–K12 states in row-major order;
normal edges are not debounced again. After context reset, held keys stay
suppressed until the library reports release. Initially undecided keys wait
nine None samples (upstream debounce 8) to exclude a latent press.

Encoder uses drivers v0.34.0 `NewQuadratureViaInterrupt` / Position, with
Precision 4, direction +1 and a reset of fractional counts on context changes.
Accumulated pending rotation is limited to 32 steps; overflow requires HELLO.
`matrixInvertDiode`, `encoderPrecision`, `encoderDirection` and joystick
calibration in `input.go` are the hardware adjustment points.

OLED uses SSD1306 v0.34.0 with TinyFont v0.6.0 TomThumb and TinyDraw v0.4.0.
The 2×3 state layout/selection/offline cross is tested with an in-memory display;
TomThumb glyphs are smaller than the old custom glyphs and physical readability
still needs acceptance. LEDs retain PIO v0.2.0, existing mapping/brightness and
are extinguished offline. Full input dependency and original font notices are
retained under `third_party/`; no workshop code was copied.

`kb02_inputonly` is mandatory; missing it fails compilation. It excludes upstream
keyboard/Vial output initialization. The static dependency list contains standard
CDC/HID mouse and no HID keyboard handler, but standard TinyGo CDCHID descriptors
include unused keyboard report items; real USB enumeration remains acceptance.

OLED/LED transfer still runs in the scan loop. IRQ Encoder capture can run during
that transfer, but matrix polling and pointer timing under display load remain
unaccepted. TinyGo 0.40.1's CDC Write buffers asynchronously and does not expose
internal TX overflow; the finite firmware queue does not prove loss-free delivery
under a stalled USB endpoint. GPIO direction, actual key numbering, encoder
polarity/detents, neutral drift, reconnects and real Herdr/Codex acceptance need
hardware. Do not flash without a separate explicit instruction.

2026-10-09: Go tests/vet and a compile-only UF2 build passed; no USB or Herdr
operation was performed. Keep the old recovery UF2 unchanged.

Slot RGB is shown only on the six agent-key LEDs at a maximum channel value of
16/255. OLED I2C runs at 400kHz and the framebuffer is transmitted only when
the six-slot panel state changes. Joystick calibration is defined once in
`input.go`.
