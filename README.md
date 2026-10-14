# tsubame-lab

ネットワーク基礎コースの練習用ツール。自分のPCの中だけで、練習用の Web サービス「つばめノート」のミニ版を動かし、
わざと故障を起こして、ブラウザや curl で観察し、直す練習に使う。

- ダウンロードした1ファイルを実行するだけで動く。インストールも管理者権限もいらない。
- 待ち受けるのは `127.0.0.1`（このPC自身）だけで、ほかの機器からは見えない。
- PCのネットワーク設定、ファイアウォール、DNS、hosts、信頼設定（ルート証明書）は変えない。
- 状態とログはホームの下の `tsubame-lab/` にだけ置き、`reset` で作り直す。要らなくなったら、実行ファイルとこのディレクトリを消せばよい。

Udemy のコース「ネットワーク基礎」の講義で、講師と一緒に使う。

## ダウンロード

[リリースページ](https://github.com/kentny/tsubame-lab/releases) から、自分のOSのファイルを1つダウンロードする。
ファイルは、このページからだけダウンロードする。

| OS | ファイル |
|---|---|
| Windows（x64） | `tsubame-lab-windows-x64.exe` |
| macOS（Apple シリコン） | `tsubame-lab-macos-apple-silicon.zip` |
| macOS（Intel） | `tsubame-lab-macos-intel.zip` |
| Linux（x64） | `tsubame-lab-linux-x64.tar.gz` |

署名を付けていないので、開くときに OS の警告が出ることがある。このページから取ったファイルなら、警告の画面で、それでも開く、という意味の選択肢を選んで進める。
ダウンロードしたファイルが壊れていないかは、同じページの `SHA256SUMS.txt` と比べて確かめられる。

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

外部のライブラリは使わない。必要なのは Git、Go（`go.mod` に書いた版以上）、make。

- Go: Mac は `brew install go`、または https://go.dev/dl/ から入れる。`go version` で版を確かめられる。
- make: Mac は Git と一緒に入る開発ツール（Xcode Command Line Tools）に入っている。Windows には入っていないので、
  下の「make を使わない」手順でビルドする。

### make を使う

```
git clone https://github.com/kentny/tsubame-lab.git
cd tsubame-lab
make test
make build
```

このディレクトリに `tsubame-lab` ができる。

| 作業 | 内容 |
|---|---|
| `make test` | 検査（`go vet`）とテスト。8000番・8080番が空いていれば、実際に起動して止める流れも確かめる |
| `make build` | このPC向けの実行ファイルを作る |
| `make dist VERSION=0.1.0` | 配布する4つのOS向けのファイルとチェックサムの一覧を `dist/` に作る（`zip` を使う） |
| `make clean` | 作ったファイルを消す |

### make を使わない（Windows など）

同じディレクトリで次を実行する。

```
go vet ./...
go test ./...
go build -o tsubame-lab.exe .
```

Mac・Linux では最後の行を `go build -o tsubame-lab .` にする。

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

- リリース: `v0.1.0` のようなタグを付けると、`.github/workflows/release.yml` がリリースページに置く。
- 仕組み: `start` は自分自身を常駐プロセスとして起動する。コマンドと常駐プロセスは、作業ディレクトリの `desired.json`
  （こうしてほしい状態）と `runtime.json`（今の状態）でやりとりする。制御用の番号では待ち受けない（受講者が待ち受けの一覧を
  見たときに、コースで扱う部品だけが見えるように）。

## ライセンス

MIT License（`LICENSE`）。
