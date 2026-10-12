# tsubame-lab

ネットワーク基礎講座の練習用ツール。自分のPCの中だけで、練習用の Web サービス「つばめノート」のミニ版を動かし、
わざと故障を起こして、ブラウザや curl で観察し、直す練習に使う。

- ダウンロードした1ファイルを実行するだけで動く。インストールも管理者権限もいらない。
- 待ち受けるのは `127.0.0.1`（このPC自身）だけで、ほかの機器からは見えない。
- PCのネットワーク設定、ファイアウォール、DNS、hosts、信頼設定（ルート証明書）は変えない。
- 状態とログはホームの下の `tsubame-lab/` にだけ置き、`reset` で作り直す。要らなくなったら、実行ファイルとこのディレクトリを消せばよい。

講座の中での使い方と仕様は、講座の演習の資料（`../../exercises/lab-tool.md`）にある。

## 中で動くもの

| 部品 | 待ち受け | 役割 |
|---|---|---|
| proxy | 127.0.0.1:8000 | 手前で受け付けて、奥の app に取り次ぐ。受け付ける名前は 127.0.0.1、localhost、notes.tsubame.test |
| app | 127.0.0.1:8080 | つばめノートのページを返す（`/`、`/notes/`、`/error`、`/slow`。それ以外は 404） |

ブラウザで開くのは `http://127.0.0.1:8000/`。

## コマンド

| コマンド | 内容 |
|---|---|
| `tsubame-lab start` | 全部の部品を起動する（止まっている部品だけを起動する） |
| `tsubame-lab stop` | 全部の部品を止めて、ツールを終わる |
| `tsubame-lab status` | 部品ごとの待ち受けと状態を表示する |
| `tsubame-lab break app-stop` | app を止める |
| `tsubame-lab break proxy-stop` | proxy を止める |
| `tsubame-lab reset` | 初期状態に戻す（作業ディレクトリを作り直して起動する） |
| `tsubame-lab logs` | ログの場所を表示する |
| `tsubame-lab version` | 版を表示する |

`start` の後は、端末を閉じても部品は動き続ける。終わるときは `stop` を実行する。

## ログ

| ファイル | 書式の例 |
|---|---|
| `logs/proxy-access.log` | `2026-10-12T09:37:31.763+09:00 client=127.0.0.1:41576 host=127.0.0.1:8000 "GET / HTTP/1.1" status=200 forward=127.0.0.1:8080 forward_status=200 rt=0.001` |
| `logs/proxy-error.log` | `2026-10-12T09:37:31.827+09:00 [error] "GET /" upstream=127.0.0.1:8080 status=502: connect: connection refused` |
| `logs/app.log` | `2026-10-12T00:37:31.763Z GET / 200 dur_ms=0`（app のログは UTC） |

## ビルドのしかた

外部のライブラリは使わない。必要なのは Go と Git だけ。

### mise を使う（おすすめ）

[mise](https://mise.jdx.dev/) が、`mise.toml` に書いた版の Go を入れ、よく使う作業をまとめて実行する。

1. mise を入れる（1回だけ）。Mac は `brew install mise`、Windows は `winget install jdx.mise`。そのほかは mise の
   ドキュメントのとおり。
2. ソースを取ってきて、Go を入れる。

   ```
   git clone https://github.com/kentny/udemy-course-design.git
   cd udemy-course-design
   git checkout network/lab-tool
   cd courses/network/tools/tsubame-lab
   mise trust
   mise install
   ```

   `mise trust` は、このディレクトリの `mise.toml` を使ってよいと mise に伝える操作（初回だけ）。

3. テストして、ビルドする。

   ```
   mise run test
   mise run build
   ```

   このディレクトリに `tsubame-lab`（Windows では `tsubame-lab.exe`）ができる。

| 作業 | 内容 |
|---|---|
| `mise run test` | 検査（`go vet`）とテスト。8000番・8080番が空いていれば、実際に起動して止める流れも確かめる |
| `mise run build` | このPC向けの実行ファイルを作る |
| `mise run dist 0.1.0` | 配布する4つのOS向けのファイルとチェックサムの一覧を `dist/` に作る（macOS・Linux で実行する。`zip` を使う） |
| `mise run clean` | 作ったファイルを消す |

### mise を使わない

Go（`mise.toml` に書いた版）を自分で入れてから、同じディレクトリで次を実行する。

```
go vet ./...
go test ./...
go build -o tsubame-lab .
```

Windows では最後の行を `go build -o tsubame-lab.exe .` にする。配布するファイルは `scripts/build.sh 0.1.0` で作る。

### 動かして確かめる

ビルドしたディレクトリで実行する（Windows では `./tsubame-lab` を `.\tsubame-lab.exe` に置き換える）。

| 操作 | ブラウザで `http://127.0.0.1:8000/` を開く（再読み込みする）と |
|---|---|
| `./tsubame-lab start` | つばめノートのページ |
| （アドレスを `/notfound` にする） | app が返す 404 のページ |
| `./tsubame-lab break app-stop` | proxy が返す 502 のページ |
| `./tsubame-lab start` | 元に戻る |
| `./tsubame-lab break proxy-stop` | ブラウザ自身の「アクセスできません」の表示 |
| `./tsubame-lab start` | 元に戻る |
| `./tsubame-lab stop` | ツールが終わる |

自分のPCでビルドしたファイルはダウンロードしたものではないので、macOS や Windows の警告は出ない想定（未確認）。

## 作る人向け

- リリース: `v0.1.0` のようなタグを付けると、`.github/workflows/release.yml` がリリースページに置く（公開リポジトリに移してから有効）。
- 仕組み: `start` は自分自身を常駐プロセスとして起動する。コマンドと常駐プロセスは、作業ディレクトリの `desired.json`
  （こうしてほしい状態）と `runtime.json`（今の状態）でやりとりする。制御用の番号では待ち受けない（受講者が待ち受けの一覧を
  見たときに、講座で扱う部品だけが見えるように）。
