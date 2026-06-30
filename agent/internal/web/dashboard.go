package web

const dashboardHTML = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Kiosk Client Administration</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #f4f6f7;
      --panel: #ffffff;
      --panel-soft: #eef2f3;
      --text: #1a2428;
      --muted: #5b696f;
      --line: #d7dfe2;
      --good: #177245;
      --warn: #916400;
      --bad: #b42318;
      --accent: #245b78;
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      background: var(--bg);
      color: var(--text);
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      line-height: 1.45;
    }

    main {
      width: min(1220px, calc(100% - 32px));
      margin: 0 auto;
      padding: 28px 0 36px;
    }

    header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 18px;
      padding-bottom: 20px;
      border-bottom: 1px solid var(--line);
    }

    h1 {
      margin: 0 0 4px;
      font-size: clamp(1.55rem, 4vw, 2.25rem);
      font-weight: 700;
      letter-spacing: 0;
    }

    h2 {
      margin: 0;
      font-size: 1rem;
      letter-spacing: 0;
    }

    p {
      margin: 0;
      color: var(--muted);
    }

    button {
      min-height: 38px;
      border: 1px solid var(--accent);
      border-radius: 6px;
      background: var(--accent);
      color: #ffffff;
      padding: 0 14px;
      font: inherit;
      cursor: pointer;
    }

    button.secondary {
      background: var(--panel);
      color: var(--accent);
    }

    button:disabled {
      cursor: wait;
      opacity: 0.7;
    }

    .actions,
    .browser-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }

    .summary {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
      margin: 20px 0;
    }

    .summary-card,
    section {
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
    }

    .summary-card {
      min-width: 0;
      padding: 14px;
    }

    .label {
      color: var(--muted);
      font-size: 0.78rem;
      font-weight: 700;
      letter-spacing: 0;
      text-transform: uppercase;
    }

    .value {
      display: block;
      min-width: 0;
      margin-top: 6px;
      overflow-wrap: anywhere;
      font-size: 1.05rem;
      font-weight: 650;
    }

    .status {
      display: inline-flex;
      align-items: center;
      min-height: 28px;
      border-radius: 999px;
      padding: 2px 10px;
      background: var(--panel-soft);
      color: var(--muted);
      font-weight: 700;
    }

    .status.healthy,
    .status.running,
    .status.ok {
      background: #e4f4ec;
      color: var(--good);
    }

    .status.degraded,
    .status.limited {
      background: #fff2cf;
      color: var(--warn);
    }

    .status.error,
    .status.stopped {
      background: #fde8e5;
      color: var(--bad);
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 14px;
    }

    section {
      min-width: 0;
      padding: 16px;
    }

    section h2 {
      padding-bottom: 12px;
      border-bottom: 1px solid var(--line);
    }

    dl {
      display: grid;
      grid-template-columns: minmax(130px, 0.72fr) minmax(0, 1fr);
      gap: 9px 14px;
      margin: 14px 0 0;
    }

    dt {
      color: var(--muted);
      font-size: 0.9rem;
    }

    dd {
      min-width: 0;
      margin: 0;
      overflow-wrap: anywhere;
      font-weight: 600;
    }

    .wide {
      grid-column: 1 / -1;
    }

    .browser-actions {
      margin-top: 14px;
    }

    .message {
      margin-top: 14px;
      color: var(--muted);
      font-weight: 650;
    }

    .message.error {
      color: var(--bad);
    }

    .history {
      width: 100%;
      margin-top: 14px;
      border-collapse: collapse;
      font-size: 0.92rem;
    }

    .history th,
    .history td {
      padding: 8px 10px;
      border-top: 1px solid var(--line);
      text-align: left;
      vertical-align: top;
      overflow-wrap: anywhere;
    }

    .history th {
      color: var(--muted);
      font-weight: 700;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        --bg: #111719;
        --panel: #182124;
        --panel-soft: #223033;
        --text: #edf2f3;
        --muted: #a7b4b9;
        --line: #2c3a3f;
        --good: #7dd9a2;
        --warn: #ffd36d;
        --bad: #ff9b92;
        --accent: #4f9fc4;
      }
    }

    @media (max-width: 900px) {
      .summary,
      .grid {
        grid-template-columns: 1fr;
      }

      header {
        flex-direction: column;
      }
    }

    @media (max-width: 560px) {
      main {
        width: min(100% - 20px, 1220px);
        padding-top: 18px;
      }

      dl {
        grid-template-columns: 1fr;
      }

      button {
        width: 100%;
      }
    }
  </style>
</head>
<body>
  <main>
    <header>
      <div>
        <h1>Kiosk Client</h1>
        <p>Lokale Administration dieser Appliance</p>
      </div>
      <div class="actions">
        <button id="refresh" type="button">Aktualisieren</button>
      </div>
    </header>

    <div class="summary" aria-live="polite">
      <div class="summary-card">
        <span class="label">Health</span>
        <span id="health-status" class="value status">laedt</span>
      </div>
      <div class="summary-card">
        <span class="label">Browser</span>
        <span id="browser-running" class="value status">laedt</span>
      </div>
      <div class="summary-card">
        <span class="label">Watchdog</span>
        <span id="watchdog-state" class="value status">laedt</span>
      </div>
      <div class="summary-card">
        <span class="label">Letzte Aktualisierung</span>
        <span id="last-refresh" class="value">laedt</span>
      </div>
    </div>

    <div id="load-message" class="message error" role="alert" hidden></div>

    <div class="grid">
      <section>
        <h2>Systemuebersicht</h2>
        <dl id="system-list"></dl>
      </section>

      <section>
        <h2>Diagnose</h2>
        <dl id="diagnostics-list"></dl>
      </section>

      <section>
        <h2>Konfiguration</h2>
        <dl id="config-list"></dl>
      </section>

      <section>
        <h2>Browser</h2>
        <dl id="browser-list"></dl>
        <div class="browser-actions">
          <button id="browser-restart" type="button">Browser neu starten</button>
          <button id="browser-reload" class="secondary" type="button">Browser reload</button>
        </div>
        <div id="browser-message" class="message" role="status"></div>
      </section>

      <section class="wide">
        <h2>Watchdog</h2>
        <dl id="watchdog-list"></dl>
        <table class="history" aria-label="Restart History">
          <thead>
            <tr>
              <th>Zeit</th>
              <th>Grund</th>
            </tr>
          </thead>
          <tbody id="restart-history"></tbody>
        </table>
      </section>

      <section class="wide">
        <h2>Metrics</h2>
        <dl id="metrics-list"></dl>
      </section>
    </div>
  </main>

  <script>
    const endpoints = {
      status: "/api/status",
      info: "/api/info",
      config: "/api/config",
      health: "/api/health",
      metrics: "/api/metrics"
    };

    const fields = {
      system: [
        ["hostname", "Hostname"],
        ["ip", "IP-Adresse"],
        ["url", "URL"],
        ["device_id", "Device-ID"],
        ["browser", "Browser"],
        ["agent_version", "Agent-Version"],
        ["os", "Betriebssystem"],
        ["kernel", "Kernel"],
        ["architecture", "Architektur"],
        ["browser_version", "Browser-Version"],
        ["browser_running", "Browserstatus"],
        ["health", "Healthstatus"]
      ],
      diagnostics: [
        ["uptime", "Uptime"],
        ["cpu_model", "CPU"],
        ["memory", "RAM"],
        ["disk", "Festplatte"],
        ["load_average", "Load Average"]
      ],
      config: [
        ["url", "URL"],
        ["device_id", "Device-ID"],
        ["browser", "Browser"],
        ["authentication", "Authentication"]
      ],
      browser: [
        ["browser_running", "Status"],
        ["browser_pid", "PID"],
        ["browser_version", "Version"],
        ["browser_path", "Pfad"],
        ["browser_cmdline", "Kommandozeile"]
      ],
      watchdog: [
        ["browser_watchdog_state", "Watchdog State"],
        ["browser_restart_count", "Restart Count"],
        ["browser_last_restart", "Last Restart"]
      ],
      metrics: [
        ["agent_uptime_seconds", "Agent-Uptime"],
        ["browser_uptime_seconds", "Browser-Uptime"],
        ["watchdog_checks", "Watchdog-Checks"],
        ["browser_restart_count", "Browser-Restarts"],
        ["http_requests_total", "HTTP-Requests"],
        ["goroutines", "Goroutines"],
        ["memory_alloc_bytes", "Go Memory Alloc"],
        ["memory_sys_bytes", "Go Memory Sys"]
      ]
    };

    const byteFields = new Set([
      "memory_total",
      "memory_available",
      "disk_total",
      "disk_available",
      "memory_alloc_bytes",
      "memory_sys_bytes"
    ]);

    const secondFields = new Set([
      "agent_uptime_seconds",
      "browser_uptime_seconds",
      "uptime"
    ]);

    const refreshButton = document.getElementById("refresh");
    const restartButton = document.getElementById("browser-restart");
    const reloadButton = document.getElementById("browser-reload");
    const loadMessage = document.getElementById("load-message");
    const browserMessage = document.getElementById("browser-message");

    let authenticationKnown = false;

    async function loadDashboard() {
      refreshButton.disabled = true;
      loadMessage.hidden = true;
      loadMessage.textContent = "";

      try {
        const names = Object.keys(endpoints);
        const responses = await Promise.all(names.map((name) => fetchJSON(endpoints[name])));
        const data = Object.fromEntries(names.map((name, index) => [name, responses[index]]));
        const view = buildView(data);

        renderSummary(view);
        renderList("system-list", fields.system, view);
        renderList("diagnostics-list", fields.diagnostics, view);
        renderList("config-list", fields.config, view);
        renderList("browser-list", fields.browser, view);
        renderList("watchdog-list", fields.watchdog, view);
        renderList("metrics-list", fields.metrics, view);
        renderHistory(data.status.browser_restart_history);
        document.getElementById("last-refresh").textContent = new Date().toLocaleString();
      } catch (error) {
        loadMessage.textContent = error.message || "Dashboard konnte nicht geladen werden.";
        loadMessage.hidden = false;
      } finally {
        refreshButton.disabled = false;
      }
    }

    async function fetchJSON(url) {
      const response = await fetch(url, { headers: { "Accept": "application/json" } });
      if (!response.ok) {
        throw new Error(url + " antwortete mit HTTP " + response.status);
      }

      return response.json();
    }

    function buildView(data) {
      const status = data.status || {};
      const info = data.info || {};
      const config = data.config || {};
      const health = data.health || {};
      const metrics = data.metrics || {};

      return {
        ...status,
        ...metrics,
        hostname: first(status.hostname, info.hostname),
        ip: status.ip,
        url: first(config.url, status.url),
        device_id: config.device_id,
        browser: first(config.browser, status.browser),
        agent_version: first(info.agent_version, status.version),
        os: joinValues([info.os_name, info.os_version]),
        kernel: first(info.kernel, status.kernel),
        architecture: first(info.architecture, status.architecture),
        browser_version: status.browser_version,
        browser_running: status.browser_running,
        health: health.status,
        uptime: status.uptime,
        cpu_model: status.cpu_model,
        memory: formatPair(status.memory_available, status.memory_total),
        disk: formatPair(status.disk_available, status.disk_total),
        load_average: status.load_average,
        authentication: authenticationKnown ? "Authentication enabled" : "not exposed"
      };
    }

    function renderSummary(view) {
      const running = view.browser_running === true;
      const watchdog = view.browser_watchdog_state || "unknown";

      setStatus("health-status", view.health || "unknown", view.health);
      setStatus("browser-running", running ? "laeuft" : "gestoppt", running ? "running" : "stopped");
      setStatus("watchdog-state", watchdog, watchdog);
    }

    function setStatus(id, text, state) {
      const element = document.getElementById(id);
      element.textContent = value(text);
      element.className = "value status " + String(state || "").toLowerCase();
    }

    function renderList(id, fieldList, data) {
      const list = document.getElementById(id);
      list.replaceChildren();

      for (const [key, label] of fieldList) {
        const term = document.createElement("dt");
        const description = document.createElement("dd");

        term.textContent = label;
        description.textContent = formatValue(key, data ? data[key] : undefined);

        list.append(term, description);
      }
    }

    function renderHistory(history) {
      const table = document.getElementById("restart-history");
      table.replaceChildren();

      if (!Array.isArray(history) || history.length === 0) {
        const row = document.createElement("tr");
        const cell = document.createElement("td");
        cell.colSpan = 2;
        cell.textContent = "-";
        row.append(cell);
        table.append(row);
        return;
      }

      for (const event of history) {
        const row = document.createElement("tr");
        row.append(
          tableCell(first(event.time, event.timestamp, event.at)),
          tableCell(first(event.reason, event.message))
        );
        table.append(row);
      }
    }

    function tableCell(raw) {
      const cell = document.createElement("td");
      cell.textContent = value(raw);
      return cell;
    }

    async function runBrowserAction(url, label) {
      setButtonsDisabled(true);
      browserMessage.className = "message";
      browserMessage.textContent = label + " wird ausgefuehrt.";

      try {
        const response = await fetch(url, { method: "POST", headers: { "Accept": "application/json" } });
        if (response.status === 401) {
          authenticationKnown = true;
          browserMessage.textContent = "Authentication enabled";
          await loadDashboard();
          return;
        }

        if (!response.ok) {
          throw new Error(label + " antwortete mit HTTP " + response.status);
        }

        browserMessage.textContent = label + " ausgefuehrt.";
        await loadDashboard();
      } catch (error) {
        browserMessage.className = "message error";
        browserMessage.textContent = error.message || label + " fehlgeschlagen.";
      } finally {
        setButtonsDisabled(false);
      }
    }

    function setButtonsDisabled(disabled) {
      restartButton.disabled = disabled;
      reloadButton.disabled = disabled;
      refreshButton.disabled = disabled;
    }

    function formatValue(key, raw) {
      if (typeof raw === "boolean") {
        return raw ? "ja" : "nein";
      }

      if (typeof raw === "number") {
        if (byteFields.has(key)) {
          return formatBytes(raw);
        }

        if (secondFields.has(key)) {
          return formatDuration(raw);
        }
      }

      return value(raw);
    }

    function formatPair(available, total) {
      if (typeof available !== "number" && typeof total !== "number") {
        return "-";
      }

      return formatBytes(available || 0) + " frei / " + formatBytes(total || 0) + " gesamt";
    }

    function formatBytes(bytes) {
      if (!Number.isFinite(bytes) || bytes <= 0) {
        return "0 B";
      }

      const units = ["B", "KiB", "MiB", "GiB", "TiB"];
      let size = bytes;
      let unit = 0;

      while (size >= 1024 && unit < units.length - 1) {
        size = size / 1024;
        unit++;
      }

      return size.toFixed(size >= 10 || unit === 0 ? 0 : 1) + " " + units[unit];
    }

    function formatDuration(rawSeconds) {
      const seconds = Number(rawSeconds);
      if (!Number.isFinite(seconds) || seconds <= 0) {
        return "0 s";
      }

      const days = Math.floor(seconds / 86400);
      const hours = Math.floor((seconds % 86400) / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      const rest = Math.floor(seconds % 60);
      const parts = [];

      if (days > 0) {
        parts.push(days + " d");
      }
      if (hours > 0) {
        parts.push(hours + " h");
      }
      if (minutes > 0) {
        parts.push(minutes + " min");
      }
      if (parts.length === 0) {
        parts.push(rest + " s");
      }

      return parts.join(" ");
    }

    function first(...values) {
      for (const item of values) {
        if (item !== null && item !== undefined && item !== "") {
          return item;
        }
      }

      return "";
    }

    function joinValues(items) {
      return items.filter((item) => item !== null && item !== undefined && item !== "").join(" ");
    }

    function value(raw) {
      if (raw === null || raw === undefined || raw === "") {
        return "-";
      }

      if (Array.isArray(raw)) {
        return raw.length === 0 ? "-" : JSON.stringify(raw);
      }

      if (typeof raw === "object") {
        return JSON.stringify(raw);
      }

      return String(raw);
    }

    refreshButton.addEventListener("click", loadDashboard);
    restartButton.addEventListener("click", () => runBrowserAction("/api/browser/restart", "Browser-Neustart"));
    reloadButton.addEventListener("click", () => runBrowserAction("/api/browser/reload", "Browser-Reload"));
    loadDashboard();
  </script>
</body>
</html>`
