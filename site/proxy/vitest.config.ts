import { cloudflareTest } from "@cloudflare/vitest-pool-workers";
import { defineConfig } from "vitest/config";

// The app Worker is bound as an auxiliary Worker named "tsubame-app", so the proxy
// is exercised through a real service binding. The stub only reproduces the
// routes the tests need (the real app lives in ../app).
// The local workerd binary can lag behind the date in wrangler.jsonc.
const LOCAL_COMPAT_DATE = "2026-08-22";

const appStub = `
export default {
  async fetch(request) {
    const path = new URL(request.url).pathname;
    if (path === "/slow") await new Promise((resolve) => setTimeout(resolve, 2000));
    return new Response("app:" + path, {
      status: path === "/notfound" ? 404 : 200,
      headers: { "X-Served-By": "tsubame-app" },
    });
  },
};
`;

export default defineConfig({
  plugins: [
    cloudflareTest({
      wrangler: { configPath: "./wrangler.jsonc" },
      miniflare: {
        compatibilityDate: LOCAL_COMPAT_DATE,
        workers: [{ name: "tsubame-app", modules: true, script: appStub, compatibilityDate: LOCAL_COMPAT_DATE }],
      },
    }),
  ],
  test: { include: ["test/**/*.test.ts"] },
});
