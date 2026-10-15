type BoxId = "resolver" | "pc" | "proxy" | "app";
type Anchor = "start" | "middle" | "end";
type Point = readonly [number, number];

interface Box {
  readonly id: BoxId;
  readonly name: string;
  readonly x: number;
  readonly y: number;
  readonly w: number;
  readonly h: number;
}

interface TextPos {
  readonly x: number;
  readonly y: number;
  readonly anchor: Anchor;
}

interface ArrowSpec {
  readonly n: 1 | 2 | 3 | 4;
  readonly from: Point;
  readonly to: Point;
  readonly both?: boolean;
  readonly label: string;
  readonly labelAt: TextPos;
  readonly noteAt: TextPos;
}

interface Layout {
  readonly cls: string;
  readonly viewBox: string;
  readonly nameSize: number;
  readonly smallSize: number;
  readonly labelSize: number;
  readonly head: number;
  readonly boxes: readonly Box[];
  readonly tagAt: (box: Box) => TextPos;
  readonly stateAt: (box: Box) => TextPos;
  readonly arrows: readonly ArrowSpec[];
}

const NAMES: Readonly<Record<BoxId, string>> = {
  resolver: "リゾルバ",
  pc: "あなたのPC",
  proxy: "取り次ぎ役",
  app: "アプリのサーバー",
};

const box = (id: BoxId, x: number, y: number, w: number, h: number): Box => ({
  id,
  name: NAMES[id],
  x,
  y,
  w,
  h,
});

const wide: Layout = {
  cls: "d-wide",
  viewBox: "0 0 1200 390",
  nameSize: 27,
  smallSize: 26,
  labelSize: 28,
  head: 20,
  boxes: [
    box("resolver", 10, 10, 240, 110),
    box("pc", 10, 220, 240, 150),
    box("proxy", 480, 220, 240, 150),
    box("app", 950, 220, 240, 150),
  ],
  tagAt: (b) => ({ x: b.x + b.w / 2, y: b.y + b.h - 20, anchor: "middle" }),
  stateAt: (b) => ({ x: b.x + 60, y: b.y + 90, anchor: "start" }),
  arrows: [
    {
      n: 1,
      from: [130, 120],
      to: [130, 220],
      both: true,
      label: "① 名前で住所を聞く",
      labelAt: { x: 160, y: 178, anchor: "start" },
      noteAt: { x: 0, y: 0, anchor: "start" },
    },
    {
      n: 2,
      from: [250, 266],
      to: [480, 266],
      label: "② リクエスト",
      labelAt: { x: 365, y: 246, anchor: "middle" },
      noteAt: { x: 365, y: 330, anchor: "middle" },
    },
    {
      n: 3,
      from: [720, 295],
      to: [950, 295],
      label: "③ 取り次ぐ",
      labelAt: { x: 835, y: 270, anchor: "middle" },
      noteAt: { x: 835, y: 352, anchor: "middle" },
    },
    {
      n: 4,
      from: [480, 322],
      to: [250, 322],
      label: "④ レスポンス",
      labelAt: { x: 365, y: 364, anchor: "middle" },
      noteAt: { x: 365, y: 330, anchor: "middle" },
    },
  ],
};

const narrow: Layout = {
  cls: "d-narrow",
  viewBox: "0 0 360 600",
  nameSize: 22,
  smallSize: 20,
  labelSize: 18,
  head: 14,
  boxes: [
    box("resolver", 10, 10, 340, 70),
    box("pc", 10, 160, 340, 70),
    box("proxy", 10, 310, 340, 70),
    box("app", 10, 460, 340, 120),
  ],
  tagAt: (b) => ({ x: b.x + b.w - 16, y: b.y + 44, anchor: "end" }),
  stateAt: (b) => ({ x: b.x + 52, y: b.y + 92, anchor: "start" }),
  arrows: [
    {
      n: 1,
      from: [40, 80],
      to: [40, 160],
      both: true,
      label: "① 名前で住所を聞く",
      labelAt: { x: 64, y: 126, anchor: "start" },
      noteAt: { x: 0, y: 0, anchor: "start" },
    },
    {
      n: 2,
      from: [40, 230],
      to: [40, 310],
      label: "② リクエスト",
      labelAt: { x: 64, y: 266, anchor: "start" },
      noteAt: { x: 64, y: 296, anchor: "start" },
    },
    {
      n: 4,
      from: [206, 310],
      to: [206, 230],
      label: "④ レスポンス",
      labelAt: { x: 220, y: 266, anchor: "start" },
      noteAt: { x: 0, y: 0, anchor: "start" },
    },
    {
      n: 3,
      from: [40, 380],
      to: [40, 460],
      label: "③ 取り次ぐ",
      labelAt: { x: 64, y: 414, anchor: "start" },
      noteAt: { x: 64, y: 446, anchor: "start" },
    },
  ],
};

const text = (cls: string, at: TextPos, size: number, body: string): string =>
  `<text class="${cls}" x="${at.x}" y="${at.y}" text-anchor="${at.anchor}" font-size="${size}">${body}</text>`;

function headPolygon(tip: Point, dir: Point, size: number): string {
  const [tx, ty] = tip;
  const [dx, dy] = dir;
  const bx = tx - dx * size;
  const by = ty - dy * size;
  const px = -dy * size * 0.6;
  const py = dx * size * 0.6;
  return `<polygon points="${tx},${ty} ${bx + px},${by + py} ${bx - px},${by - py}"/>`;
}

function arrowSvg(a: ArrowSpec, layout: Layout): string {
  const [x1, y1] = a.from;
  const [x2, y2] = a.to;
  const length = Math.hypot(x2 - x1, y2 - y1);
  const dir: Point = [(x2 - x1) / length, (y2 - y1) / length];
  const back: Point = [-dir[0], -dir[1]];
  const inset = layout.head * 0.8;
  const lx1 = a.both ? x1 + dir[0] * inset : x1;
  const ly1 = a.both ? y1 + dir[1] * inset : y1;
  const lx2 = x2 - dir[0] * inset;
  const ly2 = y2 - dir[1] * inset;
  const mx = (x1 + x2) / 2;
  const my = (y1 + y2) / 2;
  const heads = headPolygon(a.to, dir, layout.head) + (a.both ? headPolygon(a.from, back, layout.head) : "");
  const marks =
    `<g class="mark mark-ok"><circle cx="${mx}" cy="${my}" r="15"/><path d="M${mx - 7},${my} l5,6 l9,-12"/></g>` +
    `<g class="mark mark-fail"><circle cx="${mx}" cy="${my}" r="15"/><path d="M${mx - 7},${my - 7} l14,14 m0,-14 l-14,14"/></g>`;
  const notes =
    a.n === 3
      ? text("note note-fail", a.noteAt, layout.smallSize, "つながらない") +
        text("note note-timeout", a.noteAt, layout.smallSize, "返事がない")
      : "";
  return (
    `<g class="arrow arrow-${a.n}">` +
    `<line x1="${lx1}" y1="${ly1}" x2="${lx2}" y2="${ly2}"/>${heads}${marks}${notes}` +
    `${text("label", a.labelAt, layout.labelSize, a.label)}</g>`
  );
}

function boxSvg(b: Box, layout: Layout, appRunning: boolean): string {
  const name = text("name", { x: b.x + (layout.cls === "d-wide" ? b.w / 2 : 16), y: b.y + (layout.cls === "d-wide" ? 55 : 44), anchor: layout.cls === "d-wide" ? "middle" : "start" }, layout.nameSize, b.name);
  const tag = b.id === "proxy" || b.id === "app" ? text(`tag tag-${b.id}`, layout.tagAt(b), layout.smallSize, "返事を作った") : "";
  let state = "";
  if (b.id === "app") {
    const at = layout.stateAt(b);
    const dotX = at.x - 22;
    state =
      `<circle class="dot ${appRunning ? "dot-on" : "dot-off"}" cx="${dotX}" cy="${at.y - layout.smallSize * 0.35}" r="9"/>` +
      text("state", at, layout.smallSize + 2, appRunning ? "動いている" : "止まっている");
  }
  return `<g class="box box-${b.id}"><rect x="${b.x}" y="${b.y}" width="${b.w}" height="${b.h}" rx="14"/>${name}${state}${tag}</g>`;
}

function svgFor(layout: Layout, appRunning: boolean): string {
  const body = layout.boxes.map((b) => boxSvg(b, layout, appRunning)).join("") + layout.arrows.map((a) => arrowSvg(a, layout)).join("");
  return `<svg class="${layout.cls}" viewBox="${layout.viewBox}" role="img" aria-label="あなたのPCから取り次ぎ役を通って、アプリのサーバーへ届く流れの図">${body}</svg>`;
}

export function diagramHtml(appRunning: boolean): string {
  return `<div id="diagram" class="diagram" data-a1="ok" data-a2="idle" data-a3="idle" data-a4="idle" data-resp="none" data-app="${appRunning ? "running" : "stopped"}">${svgFor(wide, appRunning)}${svgFor(narrow, appRunning)}</div>`;
}
