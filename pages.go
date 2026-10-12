package main

import "fmt"

// app のページは、つばめノートの見た目でそろえる。proxy のページは、実際のプロキシと同じく
// 番号と名前だけの飾りのない見た目にする（だれが返したページかを見分けられるように）。

const appStyle = `
body { margin: 0; font-family: system-ui, -apple-system, "Segoe UI", "Hiragino Sans", "Noto Sans JP", sans-serif; color: #1f2933; background: #f7f9fb; line-height: 1.7; }
header { background: #1d4e89; color: #ffffff; padding: 16px 32px; font-size: 22px; font-weight: bold; }
main { max-width: 760px; margin: 32px auto; padding: 0 24px; }
h1 { font-size: 32px; margin: 0 0 16px; }
.note { background: #ffffff; border: 1px solid #c9d3dd; border-radius: 8px; padding: 12px 18px; margin: 12px 0; }
.note b { display: block; font-size: 18px; }
.muted { color: #3e4c59; }
footer { max-width: 760px; margin: 48px auto 24px; padding: 0 24px; color: #3e4c59; font-size: 14px; }
`

func appLayout(title, body string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="ja">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>%s</style>
</head>
<body>
<header>つばめノート（練習用）</header>
<main>
%s
</main>
<footer>tsubame-lab の app が返したページです。</footer>
</body>
</html>
`, title, appStyle, body)
}

var sampleNotes = []struct{ title, text string }{
	{"月曜の定例メモ", "議題の順番と、来週までの確認事項。"},
	{"買い出しリスト", "コピー用紙、付箋、ホワイトボード用マーカー。"},
	{"読みたい本", "ネットワークの入門書を1冊。"},
}

func appTopPage() string {
	body := "<h1>つばめノート</h1>\n<p class=\"muted\">最近のノート</p>\n"
	for _, n := range sampleNotes {
		body += fmt.Sprintf("<div class=\"note\"><b>%s</b>%s</div>\n", n.title, n.text)
	}
	return appLayout("つばめノート（練習用）", body)
}

func appNotesPage() string {
	body := "<h1>ノートの一覧</h1>\n"
	for _, n := range sampleNotes {
		body += fmt.Sprintf("<div class=\"note\"><b>%s</b>%s</div>\n", n.title, n.text)
	}
	return appLayout("ノートの一覧 - つばめノート（練習用）", body)
}

func appSlowPage() string {
	return appLayout("時間のかかるページ - つばめノート（練習用）",
		"<h1>時間のかかるページ</h1>\n<p>このページは、返事までわざと時間をかけています。</p>")
}

func appErrorPage(code int, reason, message string) string {
	return appLayout(fmt.Sprintf("%d %s", code, reason),
		fmt.Sprintf("<h1>%d %s</h1>\n<p>%s</p>", code, reason, message))
}

func proxyErrorPage(code int, reason string) string {
	return fmt.Sprintf(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>%d %s</title></head>
<body>
<center><h1>%d %s</h1></center>
<hr><center>tsubame-proxy</center>
</body>
</html>
`, code, reason, code, reason)
}
