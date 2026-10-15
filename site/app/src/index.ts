import { appErrorPage, appNotesPage, appSlowPage, appTopPage } from "./pages";

const SLOW_MS = 15_000;

function html(status: number, body: string): Response {
  return new Response(body, {
    status,
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": "no-store",
      "X-Served-By": "tsubame-app",
    },
  });
}

async function route(path: string): Promise<Response> {
  switch (path) {
    case "/":
      return html(200, appTopPage());
    case "/notes/":
      return html(200, appNotesPage());
    case "/error":
      return html(500, appErrorPage(500, "Internal Server Error", "アプリの中でエラーが起きました。"));
    case "/slow":
      await new Promise((resolve) => setTimeout(resolve, SLOW_MS));
      return html(200, appSlowPage());
    default:
      return html(404, appErrorPage(404, "Not Found", "お探しのページは見つかりませんでした。"));
  }
}

export default {
  async fetch(request: Request): Promise<Response> {
    return route(new URL(request.url).pathname);
  },
} satisfies ExportedHandler;
