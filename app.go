package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// app は共通構成の app1 の役。つばめノートのページを返す。
type app struct {
	log *lineLogger
	// slowDelay は /slow の応答までの時間。テストでは短くする。
	slowDelay time.Duration
	// silent が true の間は、接続を受け付けてリクエストを読んだまま、何も返さない（set app.silent on。Day 6）。
	silent atomic.Bool
}

func newApp(log *lineLogger) *app {
	return &app{log: log, slowDelay: 20 * time.Second}
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if a.silent.Load() {
		// 相手が待つのをやめて接続を閉じるまで、何も返さない。
		<-r.Context().Done()
		a.log.printf("%s %s %s no-response dur_ms=%d",
			time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			r.Method, r.URL.RequestURI(), time.Since(start).Milliseconds())
		return
	}
	rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	a.route(rec, r)
	// app のログは UTC（Day 12 の時刻の突き合わせで使う）。
	a.log.printf("%s %s %s %d dur_ms=%d",
		time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		r.Method, r.URL.RequestURI(), rec.status, time.Since(start).Milliseconds())
}

func (a *app) route(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		writePage(w, http.StatusOK, "no-store", appTopPage())
	case "/notes/":
		a.notes(w, r)
	case "/error":
		writePage(w, http.StatusInternalServerError, "no-store", appErrorPage(500, "Internal Server Error", "アプリの中でエラーが起きました。"))
	case "/slow":
		select {
		case <-time.After(a.slowDelay):
			writePage(w, http.StatusOK, "no-store", appSlowPage())
		case <-r.Context().Done():
		}
	default:
		writePage(w, http.StatusNotFound, "no-store", appErrorPage(404, "Not Found", "お探しのページは見つかりませんでした。"))
	}
}

// /notes/ は ETag を付け、If-None-Match が一致すれば 304 を返す（Day 7）。
func (a *app) notes(w http.ResponseWriter, r *http.Request) {
	body := appNotesPage()
	sum := sha256.Sum256([]byte(body))
	etag := fmt.Sprintf(`"%x"`, sum[:8])
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusNotModified)
		return
	}
	writePage(w, http.StatusOK, "no-cache", body)
}

func writePage(w http.ResponseWriter, status int, cacheControl, body string) {
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	// ブラウザが保存したページを出して、状態が見えなくならないようにする。
	h.Set("Cache-Control", cacheControl)
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
