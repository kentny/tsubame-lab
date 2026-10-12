//go:build windows

package main

import (
	"errors"
	"os/exec"
	"syscall"
)

const (
	detachedProcess       = 0x00000008
	createNewProcessGroup = 0x00000200
)

// 常駐プロセスを端末から切り離す（端末を閉じても動き続ける。新しい窓も開かない）。
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: detachedProcess | createNewProcessGroup,
		HideWindow:    true,
	}
}

// WSAEADDRINUSE
func isAddrInUse(err error) bool { return errors.Is(err, syscall.Errno(10048)) }

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleOutputCP = kernel32.NewProc("GetConsoleOutputCP")
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
)

// 日本語の表示が文字化けしないよう、このプログラムが動いている間だけ端末の文字コードを UTF-8 にし、
// 終わったら元に戻す。
func setupConsole() func() {
	old, _, _ := procGetConsoleOutputCP.Call()
	if old == 0 || old == 65001 {
		return func() {}
	}
	procSetConsoleOutputCP.Call(65001)
	return func() { procSetConsoleOutputCP.Call(old) }
}
