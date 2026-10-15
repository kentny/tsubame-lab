# Practice site (Cloudflare Workers spike)

A small practice website for the network course, modeled as
browser -> reverse proxy (`tsubame-proxy`) -> app server (`tsubame-app`).
Two separate Workers, so a 502/504 is produced by a different component than the app.

- `app/`: `tsubame-app`. No public URL; reachable only through the service binding.
  Routes: `/`, `/notes/`, `/error` (500), `/slow` (15 s, then 200), anything else 404.
- `proxy/`: `tsubame-proxy` on workers.dev. Forwards to the app (5 s timeout -> 504,
  app stopped or unreachable -> 502). `/_lab/` is a control page that stops/starts the app
  per learner via the `tsubame_app=stopped` cookie. The page shows the course's common diagram
  (PC, resolver, proxy, app server with numbered arrows) and lets learners open `/`, `/notes/`,
  `/notfound`, `/error`, `/slow`, animating how far the request got. Sources: `proxy/src/diagram.ts`,
  `lab-page.ts`, `lab-style.ts`, `lab-script.ts`.

## Test

```
cd site/proxy && npm test
```

Integration tests run in `@cloudflare/vitest-pool-workers` with the app bound as an auxiliary
Worker. The upstream timeout can be shortened with the `UPSTREAM_TIMEOUT_MS` env var (default 5000).

## Deploy

Log in with `npx wrangler login` first. Deploy the app before the proxy (the binding needs it):

```
cd site/app && npm install && npx wrangler deploy
cd ../proxy && npm install && npx wrangler deploy
```

A freshly deployed workers.dev URL can return 404 for up to about a minute.
