package main

import (
	"fmt"
	"net"
	"strings"
)

// setting は set で変えられる1つの項目。受講者が故障を直す操作に使う。
type setting struct {
	name  string
	def   string
	help  string
	check func(v string) (string, error) // 値を確かめて、そろえた形を返す
}

// settingOrder は set の一覧の表示順。
var settingOrder = []string{"app.bind", "app.silent"}

var settings = map[string]setting{
	"app.bind": {
		name: "app.bind", def: "127.0.0.1",
		help: "app が待ち受けるIPアドレス（0.0.0.0 にすると、この機器のすべてのIPアドレスで待ち受ける）",
		check: func(v string) (string, error) {
			ip := net.ParseIP(v)
			if ip == nil || ip.To4() == nil {
				return "", fmt.Errorf("%q は IPv4 のアドレスではありません（例: 127.0.0.1、0.0.0.0）", v)
			}
			return ip.String(), nil
		},
	},
	"app.silent": {
		name: "app.silent", def: "off",
		help: "on にすると、app は接続を受け付けたまま、何も返さない",
		check: func(v string) (string, error) {
			switch strings.ToLower(v) {
			case "on":
				return "on", nil
			case "off":
				return "off", nil
			}
			return "", fmt.Errorf("%q は使えません（on か off）", v)
		},
	},
}

// settingValue は今の値。desired に書いていなければ既定値。
func settingValue(m map[string]string, name string) string {
	if v, ok := m[name]; ok && v != "" {
		return v
	}
	return settings[name].def
}

// appAddr は app.bind の値から app の待ち受けアドレスを作る。番号は変えない。
func appAddr(m map[string]string) string {
	_, port, _ := net.SplitHostPort(defaultAddrs["app"])
	return net.JoinHostPort(settingValue(m, "app.bind"), port)
}
