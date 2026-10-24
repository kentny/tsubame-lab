package main

import (
	"io"
	"net"
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

// Day 6 の流れ（待ち受けアドレス、応答しない app、元に戻す）をたどる。
func TestDay6Flow(t *testing.T) {
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
	tl := func(args ...string) (string, error) {
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(), "TSUBAME_LAB_HOME="+home)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	must := func(args ...string) string {
		t.Helper()
		out, err := tl(args...)
		if err != nil {
			t.Fatalf("tsubame-lab %v: %v\n%s", args, err, out)
		}
		return out
	}
	t.Cleanup(func() { tl("stop") })

	if _, err := tl("set"); err == nil {
		t.Errorf("ツールが動いていないのに set が成功した")
	}
	must("start")
	if out := must("set"); !strings.Contains(out, "127.0.0.1") || !strings.Contains(out, "off") {
		t.Errorf("set の一覧: %s", out)
	}
	for _, bad := range [][]string{{"set", "app.bind", "abc"}, {"set", "app.silent", "maybe"}, {"set", "nothing", "1"}, {"set", "app.bind"}} {
		if out, err := tl(bad...); err == nil {
			t.Errorf("%v が成功した: %s", bad, out)
		}
	}

	// 待ち受けアドレス
	if out := must("set", "app.bind", "0.0.0.0"); !strings.Contains(out, "0.0.0.0:8080") {
		t.Errorf("set app.bind 0.0.0.0: %s", out)
	}
	get := func(url string, timeout time.Duration) (int, error) {
		c := &http.Client{Timeout: timeout, Transport: &http.Transport{DisableKeepAlives: true, Proxy: nil}}
		resp, err := c.Get(url)
		if err != nil {
			return 0, err
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode, nil
	}
	if s, err := get("http://127.0.0.1:8000/", 3*time.Second); s != 200 {
		t.Errorf("0.0.0.0 のとき proxy 経由: %d %v", s, err)
	}
	if out := must("break", "app-bind-local"); !strings.Contains(out, "127.0.0.1") {
		t.Errorf("break app-bind-local: %s", out)
	}
	if out := must("status"); !strings.Contains(out, "127.0.0.1:8080") {
		t.Errorf("break app-bind-local の後の status: %s", out)
	}

	// 応答しない app: 接続はできるが、待っても何も返らない
	must("break", "app-silent")
	if out := must("status"); !strings.Contains(out, "応答しない") {
		t.Errorf("app-silent の後の status: %s", out)
	}
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("app-silent でも接続はできるはず: %v", err)
	}
	conn.SetDeadline(time.Now().Add(time.Second))
	io.WriteString(conn, "GET / HTTP/1.1\r\nHost: 127.0.0.1:8080\r\n\r\n")
	if n, err := conn.Read(make([]byte, 1)); n != 0 || err == nil {
		t.Errorf("app-silent なのに応答があった: n=%d err=%v", n, err)
	}
	conn.Close()
	must("set", "app.silent", "off")
	if s, err := get("http://127.0.0.1:8080/", 3*time.Second); s != 200 {
		t.Errorf("app.silent off の後: %d %v", s, err)
	}

	// reset で既定値に戻る
	must("set", "app.silent", "on")
	must("reset")
	if out := must("set"); !strings.Contains(out, "off") {
		t.Errorf("reset の後の set: %s", out)
	}
}
