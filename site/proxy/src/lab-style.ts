export const labStyle = `
:root { --ink: #1f2933; --muted: #4b5563; --line: #cbd5e1; --indigo: #4f46e5; --red: #dc2626; --green: #00bf63; --idle: #6b7280; }
* { box-sizing: border-box; }
body { margin: 0; font-family: system-ui, -apple-system, "Segoe UI", "Hiragino Sans", "Noto Sans JP", sans-serif; font-size: 18px; color: var(--ink); background: #f8fafc; line-height: 1.6; }
header { background: #ffffff; border-bottom: 1px solid var(--line); padding: 14px 24px; font-size: 20px; font-weight: bold; }
main { max-width: 1120px; margin: 0 auto; padding: 24px 16px 48px; }
h1 { font-size: 36px; margin: 0 0 16px; }
h2 { font-size: 24px; margin: 0 0 12px; }
.card { background: #ffffff; border: 1px solid var(--line); border-radius: 12px; padding: 20px; margin-bottom: 20px; }
.caption { margin: 8px 0 0; font-size: 18px; color: var(--muted); }
.controls { display: flex; justify-content: flex-end; align-items: center; gap: 12px; flex-wrap: wrap; margin-top: 12px; }
.controls form { margin: 0; }
button, .open-link { font: inherit; font-size: 20px; min-height: 48px; padding: 8px 22px; border-radius: 8px; cursor: pointer; }
button.plain { background: #ffffff; color: var(--ink); border: 2px solid #64748b; }
button.primary { background: var(--indigo); color: #ffffff; border: 2px solid var(--indigo); }
button.try { background: #ffffff; color: var(--indigo); border: 2px solid var(--indigo); font-family: ui-monospace, "SF Mono", Menlo, monospace; font-weight: bold; }
button.try[aria-pressed="true"] { background: var(--indigo); color: #ffffff; }
button:focus-visible, a:focus-visible { outline: 3px solid #1f2933; outline-offset: 2px; }
.try-row { display: flex; flex-wrap: wrap; gap: 12px; }
.result { display: grid; grid-template-columns: minmax(200px, 280px) 1fr; gap: 24px; align-items: center; }
.result-empty { color: var(--muted); margin: 0; }
.code { font-size: 96px; font-weight: bold; line-height: 1.05; margin: 0; }
.code.bad { color: var(--red); }
.reason { font-size: 22px; color: var(--muted); margin: 4px 0 0; }
.facts { margin: 0; display: grid; grid-template-columns: max-content 1fr; gap: 6px 20px; }
.facts dt { color: var(--muted); }
.facts dd { margin: 0; font-weight: bold; }
.waiting { font-size: 28px; font-weight: bold; color: var(--indigo); margin: 0; }
.open-link { display: inline-block; margin-top: 12px; color: var(--indigo); border: 2px solid var(--indigo); text-decoration: none; background: #ffffff; }
[hidden] { display: none !important; }
.diagram svg { width: 100%; height: auto; display: block; }
.d-narrow { display: none !important; }
.diagram text { fill: var(--ink); }
.box rect { fill: #ffffff; stroke: #64748b; stroke-width: 3; }
.box .name { font-weight: bold; }
.diagram[data-app="stopped"] .box-app rect { fill: #e5e7eb; }
.dot-on { fill: var(--green); }
.dot-off { fill: #ffffff; stroke: #4b5563; stroke-width: 3; }
.tag { display: none; fill: var(--indigo); font-weight: bold; }
.diagram[data-resp="proxy"] .tag-proxy, .diagram[data-resp="app"] .tag-app { display: block; }
.diagram[data-resp="proxy"] .box-proxy rect, .diagram[data-resp="app"] .box-app rect { stroke: var(--indigo); stroke-width: 7; }
.diagram .tag { fill: var(--indigo); }
.arrow { color: var(--idle); }
.arrow line { stroke: currentColor; stroke-width: 5; transition: stroke .3s; }
.arrow polygon { fill: currentColor; }
.arrow .label { font-weight: bold; }
.mark { display: none; }
.mark circle { fill: #ffffff; stroke: currentColor; stroke-width: 4; }
.mark path { fill: none; stroke: currentColor; stroke-width: 5; stroke-linecap: round; stroke-linejoin: round; }
.note { display: none; font-weight: bold; }
.diagram .note { fill: var(--red); }
.diagram[data-a1="ok"] .arrow-1, .diagram[data-a2="ok"] .arrow-2, .diagram[data-a3="ok"] .arrow-3, .diagram[data-a4="ok"] .arrow-4 { color: var(--green); }
.diagram[data-a1="ok"] .arrow-1 .mark-ok, .diagram[data-a2="ok"] .arrow-2 .mark-ok, .diagram[data-a3="ok"] .arrow-3 .mark-ok, .diagram[data-a4="ok"] .arrow-4 .mark-ok { display: block; }
.diagram[data-a2="active"] .arrow-2, .diagram[data-a3="active"] .arrow-3, .diagram[data-a4="active"] .arrow-4 { color: var(--indigo); }
.diagram[data-a2="active"] .arrow-2 line, .diagram[data-a3="active"] .arrow-3 line, .diagram[data-a4="active"] .arrow-4 line { stroke-dasharray: 14 10; animation: flow .6s linear infinite; }
.diagram[data-a2="fail"] .arrow-2, .diagram[data-a3="fail"] .arrow-3, .diagram[data-a3="timeout"] .arrow-3 { color: var(--red); }
.diagram[data-a2="fail"] .arrow-2 .mark-fail, .diagram[data-a3="fail"] .arrow-3 .mark-fail { display: block; }
.diagram[data-a3="timeout"] .arrow-3 line { stroke-dasharray: 14 10; }
.diagram[data-a3="fail"] .arrow-3 .note-fail, .diagram[data-a3="timeout"] .arrow-3 .note-timeout { display: block; }
@keyframes flow { to { stroke-dashoffset: -24; } }
@media (prefers-reduced-motion: reduce) {
  .arrow line { transition: none; }
  .diagram .arrow line { animation: none !important; }
}
@media (max-width: 720px) {
  main { padding: 16px 16px 40px; }
  h1 { font-size: 30px; }
  .d-wide { display: none !important; }
  .d-narrow { display: block !important; }
  .controls { justify-content: flex-start; }
  .result { grid-template-columns: 1fr; gap: 12px; }
  .code { font-size: 72px; }
}
`;
