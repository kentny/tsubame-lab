import { diagramHtml } from "./diagram";
import { labScript } from "./lab-script";
import { labStyle } from "./lab-style";

const PATHS = ["/", "/notes/", "/notfound", "/error", "/slow"] as const;

const tryButtons = PATHS.map(
  (p) => `<button type="button" class="try" data-path="${p}" aria-pressed="false">${p}</button>`,
).join("\n");

const resultPanel = `<div class="result" aria-live="polite">
<p id="result-empty" class="result-empty">上のボタンを押すと、結果がここに出ます。</p>
<p id="result-waiting" class="waiting" hidden>待っています… <span id="w-time">0.0</span> 秒</p>
<div id="result-done" class="result" hidden style="grid-column: 1 / -1">
<div><p id="r-code" class="code"></p><p id="r-reason" class="reason"></p></div>
<div>
<dl class="facts">
<dt>レスポンスを返したのは</dt><dd id="r-who"></dd>
<dt>かかった時間</dt><dd id="r-time"></dd>
<dt>届いた範囲</dt><dd id="r-where"></dd>
</dl>
<a id="r-link" class="open-link" href="/" target="_blank" rel="noopener">新しいタブで開く</a>
</div>
</div>
</div>`;

export function labPage(appRunning: boolean): string {
  return `<!doctype html>
<html lang="ja">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>練習用の操作</title>
<style>${labStyle}</style>
</head>
<body>
<header>つばめノート（練習用）</header>
<main>
<h1>練習用の操作</h1>
<section class="card">
${diagramHtml(appRunning)}
<p class="caption">① は、このページを開いたときにもう済んでいます。</p>
<div class="controls">
<form method="post" action="/_lab/app/stop"><button type="submit" class="plain">アプリを止める</button></form>
<form method="post" action="/_lab/app/start"><button type="submit" class="primary">アプリを動かす</button></form>
</div>
</section>
<section class="card">
<h2>開いてみる</h2>
<div class="try-row">
${tryButtons}
</div>
</section>
<section class="card">
${resultPanel}
</section>
</main>
<script>${labScript}</script>
</body>
</html>
`;
}
