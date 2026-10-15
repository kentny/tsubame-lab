// Plain browser script, kept as a string so the Worker can inline it into the page.
// It must not contain backticks or dollar-brace sequences.
export const labScript = String.raw`
(function () {
  var diagram = document.getElementById("diagram");
  var empty = document.getElementById("result-empty");
  var waiting = document.getElementById("result-waiting");
  var done = document.getElementById("result-done");
  var buttons = document.querySelectorAll("button.try");
  var reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var REASONS = { 200: "OK", 404: "Not Found", 500: "Internal Server Error", 502: "Bad Gateway", 504: "Gateway Timeout" };
  var WHO = { "tsubame-app": "アプリのサーバー", "tsubame-proxy": "取り次ぎ役" };
  var RESP = { "tsubame-app": "app", "tsubame-proxy": "proxy" };
  var runId = 0;
  var ticker = 0;
  var stepTimer = 0;
  var controller = null;

  function show(el, on) { el.hidden = !on; }
  function setState(a2, a3, a4, resp) {
    diagram.dataset.a2 = a2;
    diagram.dataset.a3 = a3;
    diagram.dataset.a4 = a4;
    diagram.dataset.resp = resp;
  }
  function text(id, value) { document.getElementById(id).textContent = value; }
  function seconds(ms) { return (ms / 1000).toFixed(1); }
  function stopTimers() { clearInterval(ticker); clearTimeout(stepTimer); }

  function whereStopped(status, served) {
    if (served === "tsubame-app") return "アプリのサーバーまで届き、アプリのサーバーがレスポンスを返しました。";
    if (status === 502) return "③ で止まりました。取り次ぎ役はアプリのサーバーにつなげず、自分でレスポンスを作って返しました。";
    if (status === 504) return "③ で、アプリのサーバーから返事が来ませんでした。取り次ぎ役は待ちきれず、自分でレスポンスを作って返しました。";
    return "取り次ぎ役がレスポンスを返しました。";
  }

  function showResult(path, status, served, elapsed) {
    var proxyMade = served === "tsubame-proxy";
    var a3 = "ok";
    if (proxyMade && status === 502) a3 = "fail";
    else if (proxyMade && status === 504) a3 = "timeout";
    else if (proxyMade) a3 = "idle";
    setState("ok", a3, "ok", RESP[served] || "none");
    var code = document.getElementById("r-code");
    code.textContent = String(status);
    code.className = "code" + (status === 502 || status === 504 ? " bad" : "");
    text("r-reason", REASONS[status] || "");
    text("r-who", WHO[served] || "不明");
    text("r-time", seconds(elapsed) + " 秒");
    text("r-where", whereStopped(status, served));
    var link = document.getElementById("r-link");
    link.href = path;
    show(waiting, false);
    show(done, true);
  }

  function showNetworkError(elapsed) {
    setState("fail", "idle", "idle", "none");
    var code = document.getElementById("r-code");
    code.textContent = "×";
    code.className = "code bad";
    text("r-reason", "レスポンスなし");
    text("r-who", "なし");
    text("r-time", seconds(elapsed) + " 秒");
    text("r-where", "② で止まりました。リクエストが取り次ぎ役に届きませんでした。");
    show(waiting, false);
    show(done, true);
  }

  function run(path, button) {
    runId += 1;
    var mine = runId;
    if (controller) controller.abort();
    controller = new AbortController();
    stopTimers();
    buttons.forEach(function (b) { b.setAttribute("aria-pressed", b === button ? "true" : "false"); });
    var started = performance.now();
    show(empty, false);
    show(done, false);
    show(waiting, true);
    text("w-time", "0.0");
    setState("active", "idle", "idle", "none");
    stepTimer = setTimeout(function () { if (mine === runId) setState("ok", "active", "idle", "none"); }, reduceMotion ? 0 : 500);
    ticker = setInterval(function () { text("w-time", seconds(performance.now() - started)); }, 100);
    fetch(path, { cache: "no-store", signal: controller.signal })
      .then(function (res) {
        if (mine !== runId) return;
        var elapsed = performance.now() - started;
        stopTimers();
        showResult(path, res.status, res.headers.get("X-Served-By"), elapsed);
        if (res.body) res.body.cancel();
      })
      .catch(function (err) {
        if (mine !== runId || (err && err.name === "AbortError")) return;
        stopTimers();
        showNetworkError(performance.now() - started);
      });
  }

  buttons.forEach(function (button) {
    button.addEventListener("click", function () { run(button.dataset.path, button); });
  });
})();
`;
