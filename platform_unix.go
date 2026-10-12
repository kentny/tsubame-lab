//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// 常駐プロセスを端末から切り離す（端末を閉じても動き続ける）。
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func isAddrInUse(err error) bool { return false }

func setupConsole() func() { return func() {} }
