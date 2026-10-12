#!/bin/sh
# 配布するファイルを dist/ に作る。使い方: scripts/build.sh <版>（例: scripts/build.sh 0.1.0）
# Windows は .exe をそのまま、macOS は zip、Linux は tar.gz に入れる（ブラウザでダウンロードした実行ファイルに
# 実行できる印が付かないため）。最後にチェックサムの一覧 SHA256SUMS.txt を作る。
set -eu
version=${1:?版を指定してください（例: 0.1.0）}
cd "$(dirname "$0")/.."
rm -rf dist
mkdir -p dist/work
ldflags="-s -w -X main.version=$version"

build() {
	goos=$1 goarch=$2 out=$3
	CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -trimpath -ldflags "$ldflags" -o "$out" .
}

build windows amd64 dist/tsubame-lab-windows-x64.exe

for arch in arm64 amd64; do
	name=apple-silicon
	[ "$arch" = amd64 ] && name=intel
	rm -rf dist/work/*
	build darwin $arch dist/work/tsubame-lab
	(cd dist/work && zip -q -X "../tsubame-lab-macos-$name.zip" tsubame-lab)
done

rm -rf dist/work/*
build linux amd64 dist/work/tsubame-lab
tar -C dist/work -czf dist/tsubame-lab-linux-x64.tar.gz tsubame-lab

rm -rf dist/work
(cd dist && sha256sum tsubame-lab-* > SHA256SUMS.txt)
ls -l dist
