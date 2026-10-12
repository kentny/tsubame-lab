package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// lineLogger はログに1行ずつ追記する。日付などの接頭辞は呼ぶ側で付ける（部品ごとに書式が違うため）。
type lineLogger struct {
	mu sync.Mutex
	w  io.Writer
}

func openLog(path string) (*lineLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &lineLogger{w: f}, nil
}

func (l *lineLogger) printf(format string, args ...any) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.w, format+"\n", args...)
}
