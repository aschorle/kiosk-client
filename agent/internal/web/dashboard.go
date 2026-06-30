package web

const dashboardHTML = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>kiosk-client Konfiguration</title>
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
      width: min(980px, calc(100% - 32px));
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
      font-size: 1.08rem;
      letter-spacing: 0;
    }

    p {
      margin: 0;
      color: var(--muted);
    }

    section,
    details {
      margin-top: 16px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
      padding: 18px;
    }

    section h2 {
      padding-bottom: 12px;
      border-bottom: 1px solid var(--line);
    }

    button {
      min-height: 42px;
      border: 1px solid var(--accent);
      border-radius: 6px;
      background: var(--accent);
      color: #ffffff;
      padding: 0 16px;
      font: inherit;
      font-weight: 700;
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

    form {
      display: grid;
      gap: 12px;
      margin-top: 16px;
    }

    label {
      color: var(--muted);
      font-weight: 700;
    }

    input {
      width: 100%;
      min-height: 56px;
      border: 1px solid var(--line);
      border-radius: 6px;
      background: var(--panel);
      color: var(--text);
      padding: 0 14px;
      font: inherit;
      font-size: 1.12rem;
    }

    input:focus {
      outline: 2px solid var(--accent);
      outline-offset: 2px;
    }

    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 12px;
      margin-top: 14px;
    }

    .system-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .tile {
      min-width: 0;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel-soft);
      padding: 12px;
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
      font-size: 1.02rem;
      font-weight: 650;
    }

    .status {
      display: inline-flex;
      align-items: center;
      min-height: 28px;
      border-radius: 999px;
      padding: 2px 10px;
      background: var(--panel);
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

    .message {
      min-height: 24px;
      margin-top: 12px;
      color: var(--muted);
      font-weight: 650;
    }

    .message.error {
      color: var(--bad);
    }

    .message.ok {
      color: var(--good);
    }

    summary {
      cursor: pointer;
      font-size: 1.08rem;
      font-weight: 700;
    }

    dl {
      display: grid;
      grid-template-columns: minmax(160px, 0.55fr) minmax(0, 1fr);
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

    @media (max-width: 780px) {
      header,
      .actions {
        flex-direction: column;
      }

      .grid,
      .system-grid {
        grid-template-columns: 1fr;
      }
    }

    @media (max-width: 560px) {
      main {
        width: min(100% - 20px, 980px);
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
        <h1>kiosk-client</h1>
        <p>Lokale Client-Konfiguration</p>
      </div>
      <button id="refresh" class="secondary" type="button">Aktualisieren</button>
    </header>

    <div id="load-message" class="message error" role="alert" hidden></div>

    <section>
      <h2>Kiosk-Konfiguration</h2>
      <form id="config-form">
        <label for="url-input">URL</label>
        <input id="url-input" name="url" type="url" placeholder="https://example.org" autocomplete="url">
        <div class="actions">
          <button id="save-config" type="submit">Speichern</button>
        </div>
      </form>
      <div id="config-message" class="message" role="status"></div>
    </section>

    <section>
      <h2>Browser</h2>
      <div class="actions">
        <button id="browser-reload" class="secondary" type="button">Reload</button>
        <button id="browser-restart" type="button">Neustart</button>
      </div>
      <div id="browser-message" class="message" role="status"></div>
    </section>

    <section>
      <h2>Status</h2>
      <div class="grid">
        <div class="tile">
          <span class="label">Browser</span>
          <span id="browser-status" class="value status">laedt</span>
        </div>
        <div class="tile">
          <span class="label">Watchdog</span>
          <span id="watchdog-status" class="value status">laedt</span>
        </div>
        <div class="tile">
          <span class="label">Health</span>
          <span id="health-status" class="value status">laedt</span>
        </div>
        <div class="tile">
          <span class="label">Hostname</span>
          <span id="hostname" class="value">laedt</span>
        </div>
        <div class="tile">
          <span class="label">IP-Adresse</span>
          <span id="ip-address" class="value">laedt</span>
        </div>
      </div>
    </section>

    <section>
      <h2>System</h2>
      <div class="grid system-grid">
        <div class="tile">
          <span class="label">Version</span>
          <span id="version" class="value">laedt</span>
        </div>
        <div class="tile">
          <span class="label">Kernel</span>
          <span id="kernel" class="value">laedt</span>
        </div>
        <div class="tile">
          <span class="label">RAM</span>
          <span id="memory" class="value">laedt</span>
        </div>
        <div class="tile">
          <span class="label">CPU</span>
          <span id="cpu" class="value">laedt</span>
        </div>
        <div class="tile">
          <span class="label">Load</span>
          <span id="load" class="value">laedt</span>
        </div>
      </div>
    </section>

    <details>
      <summary>Erweiterte Diagnose</summary>
      <dl id="diagnostics-list"></dl>
      <table class="history" aria-label="Restart History">
        <thead>
          <tr>
            <th>Zeit</th>
            <th>Grund</th>
          </tr>
        </thead>
        <tbody id="restart-history"></tbody>
      </table>
    </details>
  </main>

  <script>
    const endpoints = {
      status: "/api/status",
      info: "/api/info",
      config: "/api/config",
      health: "/api/health",
      metrics: "/api/metrics"
    };

    const diagnosticsFields = [
      ["url", "Aktuelle URL"],
      ["device_id", "Device-ID"],
      ["browser", "Browser"],
      ["browser_pid", "Browser PID"],
      ["browser_version", "Browser-Version"],
      ["browser_path", "Browser-Pfad"],
      ["browser_cmdline", "Browser Commandline"],
      ["browser_restart_count", "Browser-Restarts"],
      ["browser_last_restart", "Last Restart"],
      ["uptime", "System-Uptime"],
      ["architecture", "Architektur"],
      ["os", "Betriebssystem"],
      ["go_version", "Go Runtime"],
      ["build_time", "Build-Zeit"],
      ["git_commit", "Git-Commit"],
      ["agent_uptime_seconds", "Agent-Uptime"],
      ["browser_uptime_seconds", "Browser-Uptime"],
      ["watchdog_checks", "Watchdog-Checks"],
      ["http_requests_total", "HTTP-Requests"],
      ["goroutines", "Goroutines"],
      ["memory_alloc_bytes", "Go Memory Alloc"],
      ["memory_sys_bytes", "Go Memory Sys"]
    ];

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
    const configForm = document.getElementById("config-form");
    const urlInput = document.getElementById("url-input");
    const saveButton = document.getElementById("save-config");
    const restartButton = document.getElementById("browser-restart");
    const reloadButton = document.getElementById("browser-reload");
    const loadMessage = document.getElementById("load-message");
    const configMessage = document.getElementById("config-message");
    const browserMessage = document.getElementById("browser-message");

    let currentConfig = { url: "", device_id: "", browser: "chromium" };

    async function loadDashboard() {
      refreshButton.disabled = true;
      loadMessage.hidden = true;
      loadMessage.textContent = "";

      try {
        const names = Object.keys(endpoints);
        const responses = await Promise.all(names.map((name) => fetchJSON(endpoints[name])));
        const data = Object.fromEntries(names.map((name, index) => [name, responses[index]]));
        const view = buildView(data);

        currentConfig = {
          url: data.config.url || "",
          device_id: data.config.device_id || "",
          browser: data.config.browser || "chromium"
        };

        urlInput.value = currentConfig.url;
        renderStatus(view);
        renderSystem(view);
        renderDiagnostics(view);
        renderHistory(data.status.browser_restart_history);
      } catch (error) {
        loadMessage.textContent = error.message || "Die Daten konnten nicht geladen werden.";
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
        url: first(config.url, status.url),
        device_id: config.device_id,
        browser: first(config.browser, status.browser),
        hostname: first(status.hostname, info.hostname),
        ip: status.ip,
        version: first(info.agent_version, status.version),
        go_version: info.go_version,
        build_time: info.build_time,
        git_commit: info.git_commit,
        os: joinValues([info.os_name, info.os_version]),
        kernel: first(info.kernel, status.kernel),
        architecture: first(info.architecture, status.architecture),
        health: health.status,
        memory: formatPair(status.memory_available, status.memory_total),
        cpu_model: status.cpu_model,
        load_average: status.load_average
      };
    }

    function renderStatus(view) {
      const running = view.browser_running === true;
      setStatus("browser-status", running ? "laeuft" : "gestoppt", running ? "running" : "stopped");
      setStatus("watchdog-status", view.browser_watchdog_state || "unknown", view.browser_watchdog_state);
      setStatus("health-status", view.health || "unknown", view.health);
      setText("hostname", view.hostname);
      setText("ip-address", view.ip);
    }

    function renderSystem(view) {
      setText("version", view.version);
      setText("kernel", view.kernel);
      setText("memory", view.memory);
      setText("cpu", view.cpu_model);
      setText("load", view.load_average);
    }

    function renderDiagnostics(view) {
      const list = document.getElementById("diagnostics-list");
      list.replaceChildren();

      for (const [key, label] of diagnosticsFields) {
        const term = document.createElement("dt");
        const description = document.createElement("dd");

        term.textContent = label;
        description.textContent = formatValue(key, view[key]);

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

    async function saveConfig(event) {
      event.preventDefault();
      setFormDisabled(true);
      configMessage.className = "message";
      configMessage.textContent = "Speichere Konfiguration.";

      const payload = {
        url: urlInput.value.trim(),
        device_id: currentConfig.device_id,
        browser: currentConfig.browser || "chromium"
      };

      try {
        const response = await fetch("/api/config", {
          method: "PUT",
          headers: {
            "Accept": "application/json",
            "Content-Type": "application/json"
          },
          body: JSON.stringify(payload)
        });

        if (response.status === 401) {
          throw new Error("Authentication enabled");
        }

        if (!response.ok) {
          const error = await response.json().catch(() => ({}));
          throw new Error(error.error || "Speichern fehlgeschlagen.");
        }

        configMessage.className = "message ok";
        configMessage.textContent = "Konfiguration gespeichert.";
        await loadDashboard();
      } catch (error) {
        configMessage.className = "message error";
        configMessage.textContent = error.message || "Speichern fehlgeschlagen.";
      } finally {
        setFormDisabled(false);
      }
    }

    async function runBrowserAction(url, label) {
      setBrowserButtonsDisabled(true);
      browserMessage.className = "message";
      browserMessage.textContent = label + " wird ausgefuehrt.";

      try {
        const response = await fetch(url, { method: "POST", headers: { "Accept": "application/json" } });
        if (response.status === 401) {
          throw new Error("Authentication enabled");
        }

        if (!response.ok) {
          const error = await response.json().catch(() => ({}));
          throw new Error(error.error || label + " fehlgeschlagen.");
        }

        browserMessage.className = "message ok";
        browserMessage.textContent = label + " ausgefuehrt.";
        await loadDashboard();
      } catch (error) {
        browserMessage.className = "message error";
        browserMessage.textContent = error.message || label + " fehlgeschlagen.";
      } finally {
        setBrowserButtonsDisabled(false);
      }
    }

    function setFormDisabled(disabled) {
      saveButton.disabled = disabled;
      urlInput.disabled = disabled;
      refreshButton.disabled = disabled;
    }

    function setBrowserButtonsDisabled(disabled) {
      restartButton.disabled = disabled;
      reloadButton.disabled = disabled;
      refreshButton.disabled = disabled;
    }

    function setStatus(id, text, state) {
      const element = document.getElementById(id);
      element.textContent = value(text);
      element.className = "value status " + String(state || "").toLowerCase();
    }

    function setText(id, text) {
      document.getElementById(id).textContent = value(text);
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
    configForm.addEventListener("submit", saveConfig);
    restartButton.addEventListener("click", () => runBrowserAction("/api/browser/restart", "Neustart"));
    reloadButton.addEventListener("click", () => runBrowserAction("/api/browser/reload", "Reload"));
    loadDashboard();
  </script>
</body>
</html>`
