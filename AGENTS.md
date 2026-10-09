# Agent instructions

- Work from one parent issue and keep the change inside this repository.
- Target `waveshare-rp2040-zero` with TinyGo 0.40.1 and Go 1.25.13.
- Keep firmware independent of Herdr semantics; emit physical input events only.
- Use USB CDC major 2 and physical KEY 1–12 / signed ENC deltas.
- Standard HID mouse is allowed; do not add HID keyboard/Vial, persistent configuration, a GUI, or a generic RPC layer.
- Keep joystick calibration values in one place and retain an inversion control.
- Bound serial input and LED brightness; avoid blocking scans and wildcard ports.
- The TinyGo workshop is reference-only; do not copy its unlicensed source.
- Build and use mock/self-tests first. Never flash without explicit user approval.
- Do not update the parent submodule pointer; the integration owner does that.

- Always build the product with `-tags kb02_inputonly`; the patched dependency must never initialize HID keyboard or Vial. Preserve its fixed SHA, patch and license.
