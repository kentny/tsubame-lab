// tsubame-lab は、ネットワーク基礎コースの練習用ツール。受講者のPCの中だけで、つばめノートのミニ版
// （proxy と app）を動かし、故障を起こして直す練習に使う。PCの設定は変えない。
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// version はリリースのときに -ldflags "-X main.version=..." で埋め込む。
var version = "0.1.0-dev"

const usage = `tsubame-lab: ネットワーク基礎コースの練習用ツール

使い方:
  tsubame-lab start              全部の部品を起動する（止まっている部品だけを起動する）
  tsubame-lab stop               全部の部品を止めて、ツールを終わる
  tsubame-lab status             部品ごとの待ち受けと状態を表示する
  tsubame-lab break <状態>       故障を起こす（app-stop、proxy-stop、app-bind-local、app-silent）
  tsubame-lab set                設定の項目と今の値を表示する
  tsubame-lab set <項目> <値>    設定を変える（例: set app.bind 0.0.0.0、set app.silent off）
  tsubame-lab reset              初期状態に戻す（作業ディレクトリを作り直して起動する）
  tsubame-lab logs               ログの場所を表示する
  tsubame-lab version            版を表示する

ブラウザで開くアドレス: http://127.0.0.1:8000/
`

func main() {
	restore := setupConsole()
	code := run(os.Args[1:])
	restore()
	os.Exit(code)
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Print(usage)
		return 0
	}
	var err error
	switch args[0] {
	case "start":
		err = cmdStart()
	case "stop":
		err = cmdStop()
	case "status":
		err = cmdStatus()
	case "break":
		if len(args) < 2 {
			err = errors.New("状態の名前を付けてください（例: tsubame-lab break app-stop）")
		} else {
			err = cmdBreak(args[1])
		}
	case "set":
		err = cmdSet(args[1:])
	case "reset":
		err = cmdReset()
	case "logs":
		err = cmdLogs()
	case "version", "--version", "-v":
		fmt.Println("tsubame-lab", version)
	case "help", "--help", "-h":
		fmt.Print(usage)
	case "__supervise":
		err = supervise()
	default:
		err = fmt.Errorf("%q というコマンドはありません\n\n%s", args[0], usage)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "tsubame-lab:", err)
		return 1
	}
	return 0
}

// breakAction は break で起こす故障。部品を止めるか、設定を変える。
type breakAction struct {
	stop    string // 止める部品
	setting string // 変える設定の項目
	value   string
	msg     string
}

// breaks は break の状態名と、起こす故障。
var breaks = map[string]breakAction{
	"app-stop":   {stop: "app", msg: "app を止めました。"},
	"proxy-stop": {stop: "proxy", msg: "proxy を止めました。"},
	// 既定と同じ状態だが、故障として明示して起こす（Day 6）。
	"app-bind-local": {setting: "app.bind", value: "127.0.0.1", msg: "app の待ち受けを 127.0.0.1 だけにしました。"},
	"app-silent":     {setting: "app.silent", value: "on", msg: "app が、接続を受け付けても何も返さないようにしました。"},
}

func cmdStart() error {
	p, err := newPaths()
	if err != nil {
		return err
	}
	if err := p.ensure(); err != nil {
		return err
	}
	r, err := update(p, func(d *desired) {
		for _, name := range componentOrder {
			d.Components[name] = stateRunning
		}
	}, true)
	if err != nil {
		return err
	}
	printStatus(r)
	if failed := failures(r); len(failed) > 0 {
		return fmt.Errorf("起動できない部品があります。\n%s", strings.Join(failed, "\n"))
	}
	fmt.Println("\nブラウザで http://127.0.0.1:8000/ を開いてください。")
	return nil
}

func cmdBreak(name string) error {
	act, ok := breaks[name]
	if !ok {
		names := make([]string, 0, len(breaks))
		for n := range breaks {
			names = append(names, n)
		}
		sort.Strings(names)
		return fmt.Errorf("%q という状態はありません（使える状態: %s）", name, strings.Join(names, "、"))
	}
	p, err := newPaths()
	if err != nil {
		return err
	}
	if !readRuntime(p).alive(time.Now()) {
		return errors.New("ツールが動いていません。先に tsubame-lab start を実行してください")
	}
	if _, err := update(p, func(d *desired) {
		if act.stop != "" {
			d.Components[act.stop] = stateStopped
		}
		if act.setting != "" {
			d.Settings[act.setting] = act.value
		}
	}, false); err != nil {
		return err
	}
	fmt.Println(act.msg)
	return nil
}

func cmdSet(args []string) error {
	p, err := newPaths()
	if err != nil {
		return err
	}
	r := readRuntime(p)
	if !r.alive(time.Now()) {
		return errors.New("ツールが動いていません。先に tsubame-lab start を実行してください")
	}
	if len(args) == 0 {
		printSettings(r)
		return nil
	}
	s, ok := settings[args[0]]
	if !ok {
		return fmt.Errorf("%q という項目はありません（使える項目: %s）", args[0], strings.Join(settingOrder, "、"))
	}
	if len(args) != 2 {
		return fmt.Errorf("値を1つ付けてください（例: tsubame-lab set %s %s）", s.name, s.def)
	}
	v, err := s.check(args[1])
	if err != nil {
		return err
	}
	r, err = update(p, func(d *desired) { d.Settings[s.name] = v }, false)
	if err != nil {
		return err
	}
	fmt.Printf("%s を %s にしました。\n", s.name, v)
	if s.name == "app.bind" {
		printStatus(r)
		if failed := failures(r); len(failed) > 0 {
			return fmt.Errorf("起動できない部品があります。\n%s", strings.Join(failed, "\n"))
		}
	}
	return nil
}

func printSettings(r *runtimeState) {
	for _, name := range settingOrder {
		fmt.Printf("  %-11s %-10s %s\n", name, settingValue(r.Settings, name), settings[name].help)
	}
}

func cmdStop() error {
	p, err := newPaths()
	if err != nil {
		return err
	}
	stopped, err := stopSupervisor(p)
	if err != nil {
		return err
	}
	if stopped {
		fmt.Println("全部の部品を止めて、ツールを終わりました。")
	} else {
		fmt.Println("ツールは動いていません。")
	}
	return nil
}

func cmdReset() error {
	p, err := newPaths()
	if err != nil {
		return err
	}
	if _, err := stopSupervisor(p); err != nil {
		return err
	}
	if err := os.RemoveAll(p.dir); err != nil {
		return fmt.Errorf("作業ディレクトリ %s を消せません: %w", p.dir, err)
	}
	fmt.Println("初期状態に戻しました。")
	return cmdStart()
}

func cmdStatus() error {
	p, err := newPaths()
	if err != nil {
		return err
	}
	r := readRuntime(p)
	if !r.alive(time.Now()) {
		fmt.Println("ツールは動いていません。")
		return nil
	}
	printStatus(r)
	return nil
}

func cmdLogs() error {
	p, err := newPaths()
	if err != nil {
		return err
	}
	fmt.Println("ログの場所:")
	for _, f := range []string{"proxy-access.log", "proxy-error.log", "app.log"} {
		fmt.Println("  " + filepath.Join(p.logs, f))
	}
	return nil
}

var stateLabels = map[string]string{
	stateRunning: "動いています",
	stateStopped: "止まっています",
	stateError:   "起動できません",
}

func printStatus(r *runtimeState) {
	for _, name := range componentOrder {
		c, ok := r.Components[name]
		if !ok {
			continue
		}
		label := stateLabels[c.State]
		if c.Note != "" && c.State == stateRunning {
			label += "（" + c.Note + "）"
		}
		fmt.Printf("  %-6s %-16s %s\n", name, c.Addr, label)
	}
}

func failures(r *runtimeState) []string {
	var out []string
	for _, name := range componentOrder {
		if c := r.Components[name]; c.State == stateError {
			out = append(out, fmt.Sprintf("  %s: %s", name, c.Error))
		}
	}
	return out
}

// update は desired を書き換えて Seq を1つ進め、常駐プロセスが反映するのを待つ。
// spawn が true で常駐プロセスが動いていなければ起動する。
func update(p paths, change func(*desired), spawn bool) (*runtimeState, error) {
	r := readRuntime(p)
	running := r.alive(time.Now())
	d := readDesired(p)
	if !running {
		// 前回の指示は捨て、常駐プロセスの初期状態（全部止まっていて、設定は既定値）から始める。
		d.Components = map[string]string{}
		d.Settings = map[string]string{}
		d.Exit = false
	}
	d.Seq++
	change(&d)
	if err := writeJSON(p.desired, d); err != nil {
		return nil, err
	}
	if !running {
		if !spawn {
			return nil, errors.New("ツールが動いていません")
		}
		if err := spawnSupervisor(p); err != nil {
			return nil, err
		}
	}
	return waitApplied(p, d.Seq)
}

func spawnSupervisor(p paths) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(filepath.Join(p.logs, "tool.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	cmd := exec.Command(exe, "__supervise")
	cmd.Stdout, cmd.Stderr = out, out
	cmd.Dir = p.dir
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ツールを起動できません: %w", err)
	}
	return cmd.Process.Release()
}

func waitApplied(p paths, seq int64) (*runtimeState, error) {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if r := readRuntime(p); r != nil && r.Applied == seq && r.alive(time.Now()) {
			return r, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil, fmt.Errorf("ツールが応答しません。ログ %s を確認してください", filepath.Join(p.logs, "tool.log"))
}

// stopSupervisor は常駐プロセスに終わるよう伝え、終わるまで待つ。動いていなければ false を返す。
func stopSupervisor(p paths) (bool, error) {
	r := readRuntime(p)
	if !r.alive(time.Now()) {
		return false, nil
	}
	d := readDesired(p)
	d.Seq++
	d.Exit = true
	if err := writeJSON(p.desired, d); err != nil {
		return false, err
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(p.runtime); errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	// 応答がなければ、プロセスを直接終わらせる。
	if proc, err := os.FindProcess(r.PID); err == nil {
		proc.Kill()
	}
	time.Sleep(300 * time.Millisecond)
	os.Remove(p.runtime)
	return true, nil
}
