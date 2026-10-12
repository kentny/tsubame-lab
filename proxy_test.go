package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestProxy(t *testing.T, upstream string) (*proxy, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var access, errs bytes.Buffer
	return newProxy(upstream, defaultServerNames, &lineLogger{w: &access}, &lineLogger{w: &errs}), &access, &errs
}

func get(t *testing.T, h http.Handler, host, path string) *http.Response {
	t.Helper()
	req := httptest.NewRequest("GET", "http://"+host+path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

func body(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func startApp(t *testing.T) string {
	t.Helper()
	a := newApp(&lineLogger{w: io.Discard})
	a.slowDelay = 2 * upstreamTimeout
	srv := httptest.NewServer(a)
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://")
}

// 何も待ち受けていない番号を得る。
func closedAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func TestProxyForwardsToApp(t *testing.T) {
	p, access, _ := newTestProxy(t, startApp(t))
	resp := get(t, p, "127.0.0.1:8000", "/")
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if b := body(t, resp); !strings.Contains(b, "つばめノート") {
		t.Errorf("トップページではない: %s", b)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q", cc)
	}
	line := access.String()
	for _, want := range []string{"host=127.0.0.1:8000", `"GET / HTTP/1.1"`, "status=200", "forward_status=200", "rt="} {
		if !strings.Contains(line, want) {
			t.Errorf("アクセスログに %q がない: %s", want, line)
		}
	}
}

func TestAppNotFoundAndError(t *testing.T) {
	p, _, _ := newTestProxy(t, startApp(t))
	resp := get(t, p, "127.0.0.1:8000", "/notfound")
	if resp.StatusCode != 404 || !strings.Contains(body(t, resp), "404 Not Found") {
		t.Errorf("/notfound: status = %d", resp.StatusCode)
	}
	resp = get(t, p, "127.0.0.1:8000", "/error")
	if resp.StatusCode != 500 || !strings.Contains(body(t, resp), "500 Internal Server Error") {
		t.Errorf("/error: status = %d", resp.StatusCode)
	}
}

func TestProxy502WhenAppStopped(t *testing.T) {
	p, access, errs := newTestProxy(t, closedAddr(t))
	resp := get(t, p, "127.0.0.1:8000", "/")
	if resp.StatusCode != 502 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	b := body(t, resp)
	if !strings.Contains(b, "502 Bad Gateway") || !strings.Contains(b, "tsubame-proxy") {
		t.Errorf("502 のページではない: %s", b)
	}
	if !strings.Contains(access.String(), "status=502") || !strings.Contains(access.String(), "forward_status=-") {
		t.Errorf("アクセスログ: %s", access.String())
	}
	if !strings.Contains(errs.String(), "connect") {
		t.Errorf("エラーログ: %s", errs.String())
	}
}

func TestProxy504WhenAppSlow(t *testing.T) {
	if testing.Short() {
		t.Skip("5秒かかる")
	}
	p, _, errs := newTestProxy(t, startApp(t))
	start := time.Now()
	resp := get(t, p, "127.0.0.1:8000", "/slow")
	if resp.StatusCode != 504 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if d := time.Since(start); d < upstreamTimeout || d > upstreamTimeout+2*time.Second {
		t.Errorf("待った時間 = %s", d)
	}
	if !strings.Contains(errs.String(), "status=504") {
		t.Errorf("エラーログ: %s", errs.String())
	}
}

func TestProxyRejectsUnknownHost(t *testing.T) {
	p, access, _ := newTestProxy(t, startApp(t))
	resp := get(t, p, "example.test", "/")
	if resp.StatusCode != 404 || !strings.Contains(body(t, resp), "tsubame-proxy") {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if !strings.Contains(access.String(), "forward=-") {
		t.Errorf("取り次いでいないはず: %s", access.String())
	}
	for _, host := range []string{"localhost:8000", "notes.tsubame.test", "LOCALHOST"} {
		if resp := get(t, p, host, "/"); resp.StatusCode != 200 {
			t.Errorf("%s: status = %d", host, resp.StatusCode)
		}
	}
}

func TestProxyRedirectsNotes(t *testing.T) {
	p, _, _ := newTestProxy(t, startApp(t))
	resp := get(t, p, "127.0.0.1:8000", "/notes")
	if resp.StatusCode != 301 || resp.Header.Get("Location") != "/notes/" {
		t.Errorf("status = %d, Location = %q", resp.StatusCode, resp.Header.Get("Location"))
	}
}

func TestNotesETag(t *testing.T) {
	addr := startApp(t)
	resp, err := http.Get("http://" + addr + "/notes/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	etag := resp.Header.Get("ETag")
	if resp.StatusCode != 200 || etag == "" {
		t.Fatalf("status = %d, ETag = %q", resp.StatusCode, etag)
	}
	req, _ := http.NewRequest("GET", "http://"+addr+"/notes/", nil)
	req.Header.Set("If-None-Match", etag)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 304 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
