# codex-zero-kb02-firmware

[English](README.md) · [システムの導入・操作](https://github.com/hoki621/codex-zero-kb02#readme)

zero-kb02の入力・表示を担当するTinyGo Firmwareです。USB CDC major 2で物理K1〜K12と符号付きEncoder差分を送ります。Joystickは標準HID mouseを使い、keyboard出力・Vial・両pushは無効です。Herdr側の意味付けはHostが担当します。

## ビルド・書き込み

親repoの`mise.toml`で固定した **TinyGo 0.40.1 / Go 1.25.13**を使います。

```sh
cd firmware
go test ./...
go vet ./...
tinygo build -o /tmp/zero-kb02.uf2 --target waveshare-rp2040-zero -tags kb02_inputonly --stack-size 8kb --size short .
```

PATHが固定版でない場合は各コマンドに`mise exec --`を付けます。`kb02_inputonly`は必須で、上流のHID keyboard・Vial初期化を除外します。ビルドだけでは実機に書き込みません。

書き込む場合はbridgeとserial monitorを閉じ、[復旧用UF2・復旧手順](docs/hardware-diagnostics.md)を確認してからBOOTSELモードへ入り、`/tmp/zero-kb02.uf2`を`RPI-RP2`へコピーします。再起動後のUSB serial名は`zero-kb02-v2`です。エージェントに依頼する場合は対象個体と書き込み操作を明示して許可してください。

## 実装と上流

- matrix/debounce: sago35/tinygo-keyboard `cf173e98f60329b7f7feba941461bb95c065c418`、MIT。vendor化したmatrix本体は無変更で、同梱patchは入力専用にするためのものです。
- ピン・LED順: sago35/keyboards zero-kb02 `4b18114b66637c5909229704c0a503bcdeccc057`、MIT。対応箇所に出典コメントを置いています。
- Encoder・SSD1306・WS2812B・描画/font: `go.mod`で固定した公開ライブラリ。
- ローカル実装: major 2 parser/session、有界queue、Joystick校正、6枠表示。workshopのソースはコピーしていません。

完全なライセンスは[third_party](third_party/README.md)と[recovery](recovery/sago35-keyboards-LICENSE.txt)に保持しています。バイナリ配布時もTomThumbの著作権表示を残してください。

HELLOとonline STATEを受けてから入力を有効にします。不正行は12秒のheartbeatを更新しません。overflow・offline・再接続では入力queueを破棄し、押していたキーは離すまで無効です。LED輝度は16/255を上限とします。Joystick校正・dead zone・反転は`input.go`、Encoder設定とmatrix極性は`main.go`にあります。

テストはprotocol・queue・入力フィルタ・メモリ上の描画を確認します。[検証記録](https://github.com/hoki621/codex-zero-kb02/blob/main/docs/verification.md)ではビルドと実機観測を分けています。TinyGoのCDC writeはbuffer方式で送達確認がなく、表示転送はscan loopを共有します。
