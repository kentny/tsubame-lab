package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// 作業ディレクトリ。設定、状態、ログはすべてここに置き、reset で作り直す。
// テストでは TSUBAME_LAB_HOME で場所を変える。
func workDir() (string, error) {
	if d := os.Getenv("TSUBAME_LAB_HOME"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "tsubame-lab"), nil
}

type paths struct {
	dir     string
	desired string // コマンドが書く「こうしてほしい」状態
	runtime string // 常駐プロセスが書く「今こうなっている」状態
	logs    string
}

func newPaths() (paths, error) {
	d, err := workDir()
	if err != nil {
		return paths{}, err
	}
	return paths{
		dir:     d,
		desired: filepath.Join(d, "desired.json"),
		runtime: filepath.Join(d, "runtime.json"),
		logs:    filepath.Join(d, "logs"),
	}, nil
}

func (p paths) ensure() error {
	return os.MkdirAll(p.logs, 0o755)
}

const (
	stateRunning = "running"
	stateStopped = "stopped"
	stateError   = "error"
)

// desired はコマンドから常駐プロセスへの指示。Seq を1つ増やして書き、常駐プロセスが
// 反映したら runtime の Applied が同じ値になる。
type desired struct {
	Seq        int64             `json:"seq"`
	Components map[string]string `json:"components"`
	// Settings は set で変えた設定（項目名 → 値）。書いていない項目は既定値。
	Settings map[string]string `json:"settings,omitempty"`
	Exit     bool              `json:"exit"`
}

type componentStatus struct {
	Addr  string `json:"addr"`
	State string `json:"state"`
	Error string `json:"error,omitempty"`
	Note  string `json:"note,omitempty"`
}

type runtimeState struct {
	PID        int                        `json:"pid"`
	Heartbeat  int64                      `json:"heartbeat"` // UnixMilli
	Applied    int64                      `json:"applied"`
	Components map[string]componentStatus `json:"components"`
	Settings   map[string]string          `json:"settings"`
}

// 常駐プロセスは heartbeatInterval ごとに runtime を書き直す。staleAfter より古ければ止まっているとみなす。
const (
	heartbeatInterval = 300 * time.Millisecond
	staleAfter        = 2 * time.Second
)

func (r *runtimeState) alive(now time.Time) bool {
	return r != nil && now.Sub(time.UnixMilli(r.Heartbeat)) < staleAfter
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	// Windows では、相手が読んでいる最中だと置き換えに失敗することがあるので少し待って繰り返す。
	var rerr error
	for i := 0; i < 20; i++ {
		if rerr = os.Rename(tmp, path); rerr == nil {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return rerr
}

func readJSON(path string, v any) error {
	var err error
	for i := 0; i < 20; i++ {
		var b []byte
		b, err = os.ReadFile(path)
		if err == nil {
			if err = json.Unmarshal(b, v); err == nil {
				return nil
			}
		}
		if errors.Is(err, os.ErrNotExist) {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}

func readRuntime(p paths) *runtimeState {
	var r runtimeState
	if err := readJSON(p.runtime, &r); err != nil {
		return nil
	}
	return &r
}

func readDesired(p paths) desired {
	d := desired{Components: map[string]string{}, Settings: map[string]string{}}
	if err := readJSON(p.desired, &d); err != nil || d.Components == nil {
		d.Components = map[string]string{}
	}
	if d.Settings == nil {
		d.Settings = map[string]string{}
	}
	return d
}
