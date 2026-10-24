package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// 部品の名前と既定の待ち受け。順番は status の表示順。
var componentOrder = []string{"proxy", "app"}

var defaultAddrs = map[string]string{
	"proxy": "127.0.0.1:8000",
	"app":   "127.0.0.1:8080",
}

var defaultServerNames = []string{"127.0.0.1", "localhost", "notes.tsubame.test"}

// component は1つの待ち受け。止めると待ち受けを閉じ、その番号への接続は断られるようになる。
type component struct {
	name    string
	addr    string
	handler http.Handler
	srv     *http.Server
	state   string
	err     string
}

func (c *component) start() {
	if c.srv != nil {
		return
	}
	ln, err := net.Listen("tcp", c.addr)
	if err != nil {
		c.state, c.err = stateError, listenErrorMessage(c.addr, err)
		return
	}
	c.srv = &http.Server{Handler: c.handler, ReadHeaderTimeout: 10 * time.Second}
	go c.srv.Serve(ln)
	c.state, c.err = stateRunning, ""
}

func (c *component) stop() {
	if c.srv != nil {
		c.srv.Close() // 待ち受けと、つながっている接続をすぐ閉じる
		c.srv = nil
	}
	c.state, c.err = stateStopped, ""
}

func listenErrorMessage(addr string, err error) string {
	if errors.Is(err, syscall.EADDRINUSE) || isAddrInUse(err) {
		return fmt.Sprintf("%s はほかのプログラムが使っています", addr)
	}
	return fmt.Sprintf("%s で待ち受けられません（%v）", addr, err)
}

// supervise は常駐プロセスの本体。desired を見て部品を起動・停止し、runtime に今の状態を書く。
func supervise() error {
	p, err := newPaths()
	if err != nil {
		return err
	}
	if err := p.ensure(); err != nil {
		return err
	}
	if r := readRuntime(p); r.alive(time.Now()) && r.PID != os.Getpid() {
		return fmt.Errorf("すでに動いています（pid %d）", r.PID)
	}

	appLog, err := openLog(filepath.Join(p.logs, "app.log"))
	if err != nil {
		return err
	}
	accessLog, err := openLog(filepath.Join(p.logs, "proxy-access.log"))
	if err != nil {
		return err
	}
	errorLog, err := openLog(filepath.Join(p.logs, "proxy-error.log"))
	if err != nil {
		return err
	}

	theApp := newApp(appLog)
	comps := map[string]*component{
		"app":   {name: "app", addr: defaultAddrs["app"], handler: theApp, state: stateStopped},
		"proxy": {name: "proxy", addr: defaultAddrs["proxy"], handler: newProxy(defaultAddrs["app"], defaultServerNames, accessLog, errorLog), state: stateStopped},
	}

	var applied int64 = -1
	current := map[string]string{}
	write := func() {
		st := runtimeState{PID: os.Getpid(), Heartbeat: time.Now().UnixMilli(), Applied: applied,
			Components: map[string]componentStatus{}, Settings: map[string]string{}}
		for name, c := range comps {
			st.Components[name] = componentStatus{Addr: c.addr, State: c.state, Error: c.err}
		}
		if theApp.silent.Load() {
			cs := st.Components["app"]
			cs.Note = "応答しない設定"
			st.Components["app"] = cs
		}
		for _, name := range settingOrder {
			st.Settings[name] = settingValue(current, name)
		}
		writeJSON(p.runtime, st)
	}

	ticker := time.NewTicker(heartbeatInterval / 3)
	defer ticker.Stop()
	lastBeat := time.Time{}
	for range ticker.C {
		if _, err := os.Stat(p.dir); err != nil {
			// 作業ディレクトリが消された。部品を止めて終わる。
			for _, c := range comps {
				c.stop()
			}
			return nil
		}
		d := readDesired(p)
		if d.Seq != applied {
			if d.Exit {
				for _, c := range comps {
					c.stop()
				}
				os.Remove(p.runtime)
				return nil
			}
			// 設定を反映する。待ち受けアドレスが変わったら、app を新しいアドレスで待ち受け直す。
			current = d.Settings
			theApp.silent.Store(settingValue(current, "app.silent") == "on")
			// 起動するかどうかは、すぐ下の Components の指示で決まる。
			if a := comps["app"]; a.addr != appAddr(current) {
				a.stop()
				a.addr = appAddr(current)
			}
			// 起動は奥（app）から、停止は手前（proxy）から。
			for _, name := range []string{"app", "proxy"} {
				if d.Components[name] == stateRunning {
					comps[name].start()
				}
			}
			for _, name := range componentOrder {
				if d.Components[name] == stateStopped {
					comps[name].stop()
				}
			}
			applied = d.Seq
			write()
			lastBeat = time.Now()
			continue
		}
		if time.Since(lastBeat) >= heartbeatInterval {
			write()
			lastBeat = time.Now()
		}
	}
	return nil
}
