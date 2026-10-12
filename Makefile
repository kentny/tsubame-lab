# よく使う作業。必要なのは Go（go.mod に書いた版以上）と make。使い方は README.md の「ビルドのしかた」。

.PHONY: help test build dist clean

help:
	@echo "make test                  検査とテスト"
	@echo "make build                 このPC向けの実行ファイル tsubame-lab を作る"
	@echo "make dist VERSION=0.1.0    配布する4つのOS向けのファイルとチェックサムを dist/ に作る"
	@echo "make clean                 作ったファイルを消す"

test:
	go vet ./...
	go test -count=1 ./...

build:
	go build -o tsubame-lab .

dist:
	@test -n "$(VERSION)" || { echo "版を指定してください（例: make dist VERSION=0.1.0）"; exit 1; }
	scripts/build.sh $(VERSION)

clean:
	rm -rf dist tsubame-lab tsubame-lab.exe
