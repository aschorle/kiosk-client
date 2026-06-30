package web

const dashboardHTML = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Kiosk Client Dashboard</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #f5f7f8;
      --panel: #ffffff;
      --panel-soft: #eef2f3;
      --text: #1b2428;
      --muted: #5d6a70;
      --line: #d8e0e3;
      --good: #177245;
      --warn: #996b00;
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
      width: min(1180px, calc(100% - 32px));
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
      font-size: clamp(1.6rem, 4vw, 2.35rem);
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

    button:disabled {
      cursor: wait;
      opacity: 0.72;
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
    .status.running {
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
      grid-template-columns: minmax(120px, 0.7fr) minmax(0, 1fr);
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

    .error {
      margin-top: 14px;
      color: var(--bad);
      font-weight: 650;
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

    @media (max-width: 860px) {
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
        width: min(100% - 20px, 1180px);
        padding-top: 18px;
      }

      dl {
        grid-template-columns: 1fr;
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
      <button id="refresh" type="button">Aktualisieren</button>
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
        <span class="label">URL</span>
        <span id="config-url" class="value">laedt</span>
      </div>
      <div class="summary-card">
        <span class="label">Letzte Aktualisierung</span>
        <span id="last-refresh" class="value">laedt</span>
      </div>
    </div>

    <div id="error" class="error" role="alert" hidden></div>

    <div class="grid">
      <section>
        <h2>Status</h2>
        <dl id="status-list"></dl>
      </section>

      <section>
        <h2>Konfiguration</h2>
        <dl id="config-list"></dl>
      </section>

      <section>
        <h2>Info</h2>
        <dl id="info-list"></dl>
      </section>

      <section>
        <h2>Metriken</h2>
        <dl id="metrics-list"></dl>
      </section>

      <section class="wide">
        <h2>Health</h2>
        <dl id="health-list"></dl>
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
      status: [
        ["hostname", "Hostname"],
        ["ip", "IP"],
        ["url", "URL"],
        ["browser", "Browser"],
        ["browser_running", "Browser laeuft"],
        ["browser_pid", "Browser PID"],
        ["browser_version", "Browser-Version"],
        ["browser_path", "Browser-Pfad"],
        ["browser_watchdog_state", "Watchdog"],
        ["browser_restart_count", "Browser-Neustarts"],
        ["browser_last_restart", "Letzter Browser-Neustart"],
        ["uptime", "System-Uptime"],
        ["kernel", "Kernel"],
        ["debian_version", "Debian-Version"],
        ["architecture", "Architektur"],
        ["cpu_model", "CPU"],
        ["memory_total", "Speicher gesamt"],
        ["memory_available", "Speicher frei"],
        ["disk_total", "Disk gesamt"],
        ["disk_available", "Disk frei"],
        ["load_average", "Load Average"]
      ],
      config: [
        ["url", "URL"],
        ["device_id", "Device ID"],
        ["browser", "Browser"]
      ],
      info: [
        ["agent_version", "Agent-Version"],
        ["go_version", "Go-Version"],
        ["hostname", "Hostname"],
        ["architecture", "Architektur"],
        ["kernel", "Kernel"],
        ["build_time", "Build-Zeit"],
        ["git_commit", "Git-Commit"],
        ["board", "Board"],
        ["os_name", "OS"],
        ["os_version", "OS-Version"]
      ],
      metrics: [
        ["agent_uptime_seconds", "Agent-Uptime"],
        ["browser_uptime_seconds", "Browser-Uptime"],
        ["watchdog_checks", "Watchdog-Checks"],
        ["browser_restart_count", "Browser-Neustarts"],
        ["http_requests_total", "HTTP-Requests"],
        ["goroutines", "Goroutines"],
        ["memory_alloc_bytes", "Go Memory Alloc"],
        ["memory_sys_bytes", "Go Memory Sys"]
      ],
      health: [
        ["status", "Status"]
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
    const errorBox = document.getElementById("error");

    async function loadDashboard() {
      refreshButton.disabled = true;
      errorBox.hidden = true;
      errorBox.textContent = "";

      try {
        const names = Object.keys(endpoints);
        const responses = await Promise.all(names.map((name) => fetchJSON(endpoints[name])));
        const data = Object.fromEntries(names.map((name, index) => [name, responses[index]]));

        renderSummary(data);
        renderList("status-list", fields.status, data.status);
        renderList("config-list", fields.config, data.config);
        renderList("info-list", fields.info, data.info);
        renderList("metrics-list", fields.metrics, data.metrics);
        renderList("health-list", fields.health, data.health);
        document.getElementById("last-refresh").textContent = new Date().toLocaleString();
      } catch (error) {
        errorBox.textContent = error.message || "Dashboard konnte nicht geladen werden.";
        errorBox.hidden = false;
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

    function renderSummary(data) {
      const health = data.health.status || "unknown";
      const running = data.status.browser_running === true;
      setStatus("health-status", health, health);
      setStatus("browser-running", running ? "laeuft" : "gestoppt", running ? "running" : "stopped");
      document.getElementById("config-url").textContent = value(data.config.url || data.status.url);
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

    function formatDuration(seconds) {
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
    loadDashboard();
  </script>
</body>
</html>`
