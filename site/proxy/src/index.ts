import { labPage, proxyErrorPage } from "./pages";

interface Env {
  APP: Fetcher;
}

const UPSTREAM_TIMEOUT_MS = 5_000;
const COOKIE_ATTRS = "Path=/; Secure; HttpOnly; SameSite=Lax";
const REASONS: Readonly<Record<number, string>> = {
  502: "Bad Gateway",
  504: "Gateway Timeout",
};

const errorResponse = (status: number): Response =>
  new Response(proxyErrorPage(status, REASONS[status] ?? "Error"), {
    status,
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": "no-store",
      "X-Served-By": "tsubame-proxy",
    },
  });

const isStopped = (request: Request): boolean =>
  /(?:^|;\s*)tsubame_app=stopped(?:;|$)/.test(request.headers.get("Cookie") ?? "");

function labResponse(request: Request, path: string): Response {
  const noStore = { "Cache-Control": "no-store", "X-Served-By": "tsubame-proxy" };
  if (path === "/_lab/" || path === "/_lab") {
    if (request.method !== "GET" && request.method !== "HEAD") return methodNotAllowed("GET, HEAD");
    return new Response(labPage(!isStopped(request)), {
      headers: { ...noStore, "Content-Type": "text/html; charset=utf-8" },
    });
  }
  if (path === "/_lab/app/stop" || path === "/_lab/app/start") {
    if (request.method !== "POST") return methodNotAllowed("POST");
    const cookie =
      path === "/_lab/app/stop"
        ? `tsubame_app=stopped; ${COOKIE_ATTRS}; Max-Age=86400`
        : `tsubame_app=; ${COOKIE_ATTRS}; Max-Age=0`;
    return new Response(null, {
      status: 303,
      headers: { ...noStore, Location: "/_lab/", "Set-Cookie": cookie },
    });
  }
  return notFound();
}

const notFound = (): Response =>
  new Response(proxyErrorPage(404, "Not Found"), {
    status: 404,
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": "no-store",
      "X-Served-By": "tsubame-proxy",
    },
  });

const methodNotAllowed = (allow: string): Response =>
  new Response(proxyErrorPage(405, "Method Not Allowed"), {
    status: 405,
    headers: {
      Allow: allow,
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": "no-store",
      "X-Served-By": "tsubame-proxy",
    },
  });

async function forward(request: Request, env: Env): Promise<Response> {
  const timeout = new Promise<"timeout">((resolve) =>
    setTimeout(() => resolve("timeout"), UPSTREAM_TIMEOUT_MS),
  );
  try {
    const result = await Promise.race([env.APP.fetch(request), timeout]);
    if (result === "timeout") return errorResponse(504);
    const headers = new Headers(result.headers);
    headers.set("Via", "1.1 tsubame-proxy");
    return new Response(result.body, { status: result.status, headers });
  } catch (error) {
    console.error("upstream fetch failed", error);
    return errorResponse(502);
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const path = new URL(request.url).pathname;
    if (path === "/_lab" || path.startsWith("/_lab/")) return labResponse(request, path);
    if (isStopped(request)) return errorResponse(502);
    return forward(request, env);
  },
} satisfies ExportedHandler<Env>;
