export const proxyErrorPage = (code: number, reason: string): string => `<!doctype html>
<html>
<head><meta charset="utf-8"><title>${code} ${reason}</title></head>
<body>
<center><h1>${code} ${reason}</h1></center>
<hr><center>tsubame-proxy</center>
</body>
</html>
`;

export function labPage(appRunning: boolean): string {
  const state = appRunning ? "動いている" : "止まっている";
  return `<!doctype html>
<html lang="ja">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>練習用の操作</title>
<style>
body { margin: 0; font-family: system-ui, -apple-system, "Segoe UI", "Hiragino Sans", "Noto Sans JP", sans-serif; color: #1f2933; background: #f7f9fb; line-height: 1.7; }
header { background: #1d4e89; color: #ffffff; padding: 16px 32px; font-size: 22px; font-weight: bold; }
main { max-width: 760px; margin: 32px auto; padding: 0 24px; }
h1 { font-size: 32px; margin: 0 0 16px; }
form { display: inline-block; margin-right: 12px; }
button { font-size: 18px; padding: 8px 20px; }
</style>
</head>
<body>
<header>つばめノート（練習用）</header>
<main>
<h1>練習用の操作</h1>
<p>いまの app の状態: <b>${state}</b></p>
<form method="post" action="/_lab/app/stop"><button type="submit">app を止める</button></form>
<form method="post" action="/_lab/app/start"><button type="submit">app を動かす</button></form>
<h2>ページを開く</h2>
<ul>
<li><a href="/">/</a></li>
<li><a href="/notfound">/notfound</a></li>
<li><a href="/error">/error</a></li>
<li><a href="/slow">/slow</a></li>
</ul>
</main>
</body>
</html>
`;
}
