# codex-zero-kb02-firmware

[English](README.md) · [システム全体の導入・操作](https://github.com/hoki621/codex-zero-kb02#readme)

zero-kb02をCodexのコントローラーにするTinyGo Firmwareです。キー・Encoder入力をMacへ送り、会話の状態をOLEDとLEDに表示します。Joystickではマウスポインターを移動できます。[Hostアプリケーション](https://github.com/hoki621/codex-zero-kb02-host)と組み合わせて使います。

## ビルド

[初期設定](https://github.com/hoki621/codex-zero-kb02#初期設定)に沿って親リポジトリをsubmoduleごと取得し、`mise install`を実行します。**TinyGo 0.40.1 / Go 1.25.13**を使います。以下は親リポジトリのルートで実行してください。

```sh
mise exec -- sh -c 'cd firmware && go test ./... && go vet ./...'
mise exec -- sh -c 'cd firmware && tinygo build -o /tmp/zero-kb02.uf2 --target waveshare-rp2040-zero -tags kb02_inputonly --stack-size 8kb --size short .'
```

出力は`/tmp/zero-kb02.uf2`です。`kb02_inputonly`は必須で、HID keyboard・Vialの初期化を除外し、キー操作をHostで扱うために使います。ビルドだけではデバイスへ書き込みません。

## 書き込み・接続

1. Host bridgeとserial monitorを閉じ、[復旧用Firmwareと手順](docs/hardware-diagnostics.md)を用意します。
2. デバイスをBOOTSELモードにします。`RPI-RP2`ドライブが表示されたら、`/tmp/zero-kb02.uf2`をコピーします。
3. 再起動後のUSB serial識別名は`zero-kb02-v2`です。[起動手順](https://github.com/hoki621/codex-zero-kb02#起動)に沿って、実際のUSBポートを指定してHostを接続します。

Hostが接続してonline状態を送るとキーが有効になります。正常なheartbeatが12秒間届かなければoffline表示になり、LEDが消灯します。再接続後は押していたキーを離してから使ってください。

## 開発

通信は[USB CDC protocol major 2](https://github.com/hoki621/codex-zero-kb02/blob/main/PROTOCOL.md)です。major 1とは互換性がありません。Joystickは標準HID mouse出力を使います。HID keyboard出力・Vial・押し込み操作は無効です。

Joystickの校正・dead zone・方向設定は`input.go`、Encoder設定とmatrix極性は`main.go`にあります。LED輝度の上限は16/255です。

テストは通信仕様・入力queue・フィルタ・メモリ上の描画を確認します。実機確認の代わりにはならないため、確認済みの範囲は[検証記録](https://github.com/hoki621/codex-zero-kb02/blob/main/docs/verification.md)を参照してください。表示転送は入力scan loopを共有し、USB CDC書き込みには送達確認がありません。

## ライブラリ・ライセンス

- キーのmatrix scan・debounce：MITライセンスの[sago35/tinygo-keyboard](https://github.com/sago35/tinygo-keyboard)。入力専用ビルドのpatchとともに同梱しています。
- ピン・LED順の対応：MITライセンスの[sago35/keyboards](https://github.com/sago35/keyboards)。対応する定義に出典コメントがあります。
- Encoder・画面・LED・font：`go.mod`で固定した公開ライブラリ。

デバイスの通信仕様、Joystickの校正、6枠の表示対応はこのプロジェクトで実装しています。使用リビジョンと完全なライセンス表記は[third_party](third_party/README.md)と[recovery](recovery/sago35-keyboards-LICENSE.txt)にあります。バイナリ配布時も上流のライセンスとTomThumb fontの著作権表示を保持してください。
