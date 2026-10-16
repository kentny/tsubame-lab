const appStyle = `
body { margin: 0; font-family: system-ui, -apple-system, "Segoe UI", "Hiragino Sans", "Noto Sans JP", sans-serif; color: #1f2933; background: #f7f9fb; line-height: 1.7; }
header { background: #1d4e89; color: #ffffff; padding: 16px 32px; font-size: 22px; font-weight: bold; }
main { max-width: 760px; margin: 32px auto; padding: 0 24px; }
h1 { font-size: 32px; margin: 0 0 16px; }
.note { background: #ffffff; border: 1px solid #c9d3dd; border-radius: 8px; padding: 12px 18px; margin: 12px 0; }
.note b { display: block; font-size: 18px; }
.muted { color: #3e4c59; }
footer { max-width: 760px; margin: 48px auto 24px; padding: 0 24px; color: #3e4c59; font-size: 14px; }
`;

export function appLayout(title: string, body: string): string {
  return `<!doctype html>
<html lang="ja">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${title}</title>
<style>${appStyle}</style>
</head>
<body>
<header>つばめノート（練習用）</header>
<main>
${body}
</main>
<footer>tsubame-lab の app が返したページです。</footer>
</body>
</html>
`;
}

const sampleNotes: readonly { title: string; text: string }[] = [
  { title: "月曜の定例メモ", text: "議題の順番と、来週までの確認事項。" },
  { title: "買い出しリスト", text: "コピー用紙、付箋、ホワイトボード用マーカー。" },
  { title: "読みたい本", text: "ネットワークの入門書を1冊。" },
];

const noteHtml = sampleNotes
  .map((n) => `<div class="note"><b>${n.title}</b>${n.text}</div>\n`)
  .join("");

export const appTopPage = (): string =>
  appLayout(
    "つばめノート（練習用）",
    `<h1>つばめノート</h1>\n<p class="muted">最近のノート</p>\n${noteHtml}`,
  );

export const appNotesPage = (): string =>
  appLayout("ノートの一覧 - つばめノート（練習用）", `<h1>ノートの一覧</h1>\n${noteHtml}`);

export const appSlowPage = (): string =>
  appLayout(
    "時間のかかるページ - つばめノート（練習用）",
    "<h1>時間のかかるページ</h1>\n<p>このページは、応答までわざと時間をかけています。</p>",
  );

export const appErrorPage = (code: number, reason: string, message: string): string =>
  appLayout(`${code} ${reason}`, `<h1>${code} ${reason}</h1>\n<p>${message}</p>`);
