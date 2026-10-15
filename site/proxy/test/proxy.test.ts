import { env } from "cloudflare:workers";
import { describe, expect, it } from "vitest";
import worker from "../src/index";

const call = (path: string, init?: RequestInit, overrides: Record<string, unknown> = {}) =>
  worker.fetch(new Request(`https://proxy.test${path}`, init) as never, { ...env, ...overrides } as never);

describe("forwarding", () => {
  it("forwards to the app and adds Via while keeping X-Served-By", async () => {
    // Given - アプリが動いている（Cookie なし）

    // When - 取り次ぎ役に / を要求する
    const res = await call("/");

    // Then - アプリの返事が 200 で返り、Via と X-Served-By が付く
    expect(res.status).toBe(200);
    expect(await res.text()).toBe("app:/");
    expect(res.headers.get("Via")).toBe("1.1 tsubame-proxy");
    expect(res.headers.get("X-Served-By")).toBe("tsubame-app");
  });

  it("passes the app's 404 through unchanged", async () => {
    // Given - アプリが動いている

    // When - アプリが 404 を返すパスを要求する
    const res = await call("/notfound");

    // Then - 404 はアプリが作った返事として通る
    expect(res.status).toBe(404);
    expect(res.headers.get("X-Served-By")).toBe("tsubame-app");
  });

  it("returns 502 from the proxy itself when the app is stopped by cookie", async () => {
    // Given - tsubame_app=stopped の Cookie を持っている

    // When - / を要求する
    const res = await call("/", { headers: { Cookie: "a=1; tsubame_app=stopped" } });

    // Then - アプリを呼ばずに取り次ぎ役が 502 を返す
    expect(res.status).toBe(502);
    expect(res.headers.get("X-Served-By")).toBe("tsubame-proxy");
    expect(res.headers.get("Via")).toBeNull();
    expect(await res.text()).toContain("tsubame-proxy");
  });

  it("returns 504 from the proxy when the app does not answer in time", async () => {
    // Given - アプリが 2 秒かかる /slow で、取り次ぎ役の待ち時間を 100 ms にする

    // When - /slow を要求する
    const res = await call("/slow", undefined, { UPSTREAM_TIMEOUT_MS: "100" });

    // Then - 取り次ぎ役が 504 を返す
    expect(res.status).toBe(504);
    expect(res.headers.get("X-Served-By")).toBe("tsubame-proxy");
  });

  it("returns 502 when the binding itself fails", async () => {
    // Given - APP バインディングが例外を投げる

    // When - / を要求する
    const broken = { fetch: () => Promise.reject(new Error("boom")) };
    const res = await call("/", undefined, { APP: broken });

    // Then - 取り次ぎ役が 502 を返す
    expect(res.status).toBe(502);
    expect(res.headers.get("X-Served-By")).toBe("tsubame-proxy");
  });
});

describe("lab controls", () => {
  it("sets the stopped cookie on POST /_lab/app/stop", async () => {
    // Given - 何も設定されていない

    // When - 停止ボタンの POST を送る
    const res = await call("/_lab/app/stop", { method: "POST" });

    // Then - 303 で /_lab/ へ戻り、stopped の Cookie が付く
    expect(res.status).toBe(303);
    expect(res.headers.get("Location")).toBe("/_lab/");
    expect(res.headers.get("Set-Cookie")).toContain("tsubame_app=stopped");
  });

  it("clears the cookie on POST /_lab/app/start", async () => {
    // Given - 停止中の Cookie を持っている

    // When - 開始ボタンの POST を送る
    const res = await call("/_lab/app/start", { method: "POST", headers: { Cookie: "tsubame_app=stopped" } });

    // Then - 303 で /_lab/ へ戻り、Cookie が消える（Max-Age=0）
    expect(res.status).toBe(303);
    expect(res.headers.get("Location")).toBe("/_lab/");
    expect(res.headers.get("Set-Cookie")).toMatch(/tsubame_app=;.*Max-Age=0/);
  });

  it.each(["/_lab/app/stop", "/_lab/app/start"])("rejects GET %s with 405", async (path) => {
    // Given - 操作用のパス

    // When - GET で要求する
    const res = await call(path);

    // Then - 405 で、許可されるのは POST だけ
    expect(res.status).toBe(405);
    expect(res.headers.get("Allow")).toBe("POST");
  });

  it("serves /_lab/ with no-store and the diagram", async () => {
    // Given - アプリが動いている

    // When - /_lab/ を開く
    const res = await call("/_lab/");
    const html = await res.text();

    // Then - 200 で no-store、図と 4 つの箱と開いてみるボタンが含まれる
    expect(res.status).toBe(200);
    expect(res.headers.get("Cache-Control")).toBe("no-store");
    expect(html).toContain('id="diagram"');
    for (const name of ["あなたのPC", "リゾルバ", "取り次ぎ役", "アプリのサーバー", "動いている"]) {
      expect(html).toContain(name);
    }
    expect(html).toContain('data-path="/slow"');
  });

  it("shows the stopped state on /_lab/ when the cookie is set", async () => {
    // Given - 停止中の Cookie を持っている

    // When - /_lab/ を開く
    const res = await call("/_lab/", { headers: { Cookie: "tsubame_app=stopped" } });

    // Then - 図に「止まっている」と出る
    expect(await res.text()).toContain("止まっている");
  });
});
