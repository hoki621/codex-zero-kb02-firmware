module github.com/hoki621/codex-zero-kb02-firmware

go 1.25.0

toolchain go1.25.13

require (
	github.com/sago35/tinygo-keyboard v0.0.0
	github.com/tinygo-org/pio v0.2.0
	tinygo.org/x/drivers v0.34.0
	tinygo.org/x/tinydraw v0.4.0
	tinygo.org/x/tinyfont v0.6.0
)

require (
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
)

replace github.com/sago35/tinygo-keyboard => ./third_party/tinygo-keyboard
