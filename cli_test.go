package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 実際にビルドしたファイルで、Day 1 の演習の流れ（start → break → start → reset → stop）をたどる。
// 既定の番号（8000、8080）を使うので、それらが空いているときだけ動かす。
func TestDay1Flow(t *testing.T) {
	if testing.Short() {
		t.Skip("実際に待ち受ける")
	}
	for _, addr := range defaultAddrs {
		if c, err := http.Get("http://" + addr + "/"); err == nil {
			c.Body.Close()
			t.Skipf("%s が使われている", addr)
		}
	}
	bin := filepath.Join(t.TempDir(), "tsubame-lab")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	home := filepath.Join(t.TempDir(), "home")
	tl := func(args ...string) string {
		t.Helper()
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(), "TSUBAME_LAB_HOME="+home)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("tsubame-lab %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	t.Cleanup(func() {
		cmd := exec.Command(bin, "stop")
		cmd.Env = append(os.Environ(), "TSUBAME_LAB_HOME="+home)
		cmd.Run()
	})
	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
	status := func() int {
		t.Helper()
		resp, err := client.Get("http://127.0.0.1:8000/")
		if err != nil {
			return 0 // 接続できず、応答が返らない
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}

	tl("start")
	if s := status(); s != 200 {
		t.Fatalf("start の後: %d", s)
	}
	tl("break", "app-stop")
	if s := status(); s != 502 {
		t.Errorf("app-stop の後: %d", s)
	}
	tl("start")
	if s := status(); s != 200 {
		t.Errorf("start で直した後: %d", s)
	}
	tl("break", "proxy-stop")
	if s := status(); s != 0 {
		t.Errorf("proxy-stop の後: %d", s)
	}
	if out := tl("status"); !strings.Contains(out, "止まっています") {
		t.Errorf("status: %s", out)
	}
	tl("reset")
	if s := status(); s != 200 {
		t.Errorf("reset の後: %d", s)
	}
	if _, err := os.Stat(filepath.Join(home, "logs", "proxy-access.log")); err != nil {
		t.Errorf("reset の後のログ: %v", err)
	}
	tl("stop")
	if s := status(); s != 0 {
		t.Errorf("stop の後: %d", s)
	}
	if out := tl("status"); !strings.Contains(out, "動いていません") {
		t.Errorf("stop の後の status: %s", out)
	}
}
