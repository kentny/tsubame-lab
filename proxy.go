package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// proxy は共通構成の proxy1 の役。受け付けた要求を app へ取り次ぎ、アクセスログとエラーログを書く。
type proxy struct {
	upstream    string   // 取り次ぎ先（host:port）
	serverNames []string // 受け付ける Host
	access      *lineLogger
	errlog      *lineLogger
	rp          *httputil.ReverseProxy
}

// 上流を待つ時間（変更不可。Day 8 の 504）。
const upstreamTimeout = 5 * time.Second

type proxyCtxKey struct{}

// 1つの要求について、取り次いだかどうかと上流の応答の番号を記録する。
type forwardInfo struct {
	forwarded bool
	status    int
}

func newProxy(upstream string, serverNames []string, access, errlog *lineLogger) *proxy {
	p := &proxy{upstream: upstream, serverNames: serverNames, access: access, errlog: errlog}
	transport := &http.Transport{
		Proxy:                 nil, // 受講者のPCの HTTP_PROXY を使わない
		DialContext:           (&net.Dialer{Timeout: upstreamTimeout}).DialContext,
		ResponseHeaderTimeout: upstreamTimeout,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
	}
	target := &url.URL{Scheme: "http", Host: upstream}
	p.rp = &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.Host = pr.In.Host
			pr.SetXForwarded()
		},
		ModifyResponse: func(resp *http.Response) error {
			if fi, ok := resp.Request.Context().Value(proxyCtxKey{}).(*forwardInfo); ok {
				fi.status = resp.StatusCode
			}
			return nil
		},
		ErrorHandler: p.upstreamError,
	}
	return p
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	rec.Header().Set("Server", "tsubame-proxy")
	fi := &forwardInfo{}

	switch {
	case !p.acceptsHost(r.Host):
		// 受け付けない名前には proxy 自身が 404 を返す（Day 7）。
		writePage(rec, http.StatusNotFound, "no-store", proxyErrorPage(404, "Not Found"))
	case r.URL.Path == "/notes":
		// /notes から /notes/ への 301 も proxy 自身が返す（Day 7）。
		loc := "/notes/"
		if r.URL.RawQuery != "" {
			loc += "?" + r.URL.RawQuery
		}
		rec.Header().Set("Location", loc)
		writePage(rec, http.StatusMovedPermanently, "no-store", proxyErrorPage(301, "Moved Permanently"))
	default:
		fi.forwarded = true
		p.rp.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), proxyCtxKey{}, fi)))
	}

	forward, forwardStatus := "-", "-"
	if fi.forwarded {
		forward = p.upstream
		if fi.status != 0 {
			forwardStatus = strconv.Itoa(fi.status)
		}
	}
	host := r.Host
	if host == "" {
		host = "-"
	}
	p.access.printf("%s client=%s host=%s \"%s %s %s\" status=%d forward=%s forward_status=%s rt=%.3f",
		localTime(time.Now()), r.RemoteAddr, host, r.Method, r.URL.RequestURI(), r.Proto,
		rec.status, forward, forwardStatus, time.Since(start).Seconds())
}

func (p *proxy) acceptsHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	for _, n := range p.serverNames {
		if host == n {
			return true
		}
	}
	return false
}

// 上流に取り次げなかったとき。待ちきれなければ 504、それ以外（接続を断られたなど）は 502。
func (p *proxy) upstreamError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, context.Canceled) {
		// ブラウザ側が待つのをやめた。応答は誰にも届かない。
		w.WriteHeader(499)
		return
	}
	status, reason := http.StatusBadGateway, "Bad Gateway"
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		status, reason = http.StatusGatewayTimeout, "Gateway Timeout"
	}
	p.errlog.printf("%s [error] %q upstream=%s status=%d: %s",
		localTime(time.Now()), r.Method+" "+r.URL.RequestURI(), p.upstream, status, describeUpstreamError(err))
	writePage(w, status, "no-store", proxyErrorPage(status, reason))
}

// エラーログに書く理由。Go の内部の言い回しを避け、OS が返した理由をそのまま書く。
func describeUpstreamError(err error) string {
	var oe *net.OpError
	if errors.As(err, &oe) {
		op := oe.Op
		if op == "dial" {
			op = "connect"
		}
		var se *os.SyscallError
		if errors.As(oe.Err, &se) {
			return fmt.Sprintf("%s: %v", op, se.Err)
		}
		return fmt.Sprintf("%s: %v", op, oe.Err)
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return fmt.Sprintf("no response within %s", upstreamTimeout)
	}
	return err.Error()
}

// proxy のログはPCの時刻帯（オフセットつき）。
func localTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}
