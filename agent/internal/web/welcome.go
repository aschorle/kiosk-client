package web

const welcomeHTML = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>kiosk-client - Willkommen</title>
  <style>
    :root {
      color-scheme: light dark;
      --bg: #f4f6f7;
      --panel: #ffffff;
      --text: #1a2428;
      --muted: #5b696f;
      --line: #d7dfe2;
      --accent: #245b78;
    }

    * {
      box-sizing: border-box;
    }

    body {
      min-height: 100vh;
      margin: 0;
      display: grid;
      place-items: center;
      background: var(--bg);
      color: var(--text);
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      line-height: 1.45;
    }

    main {
      width: min(760px, calc(100% - 32px));
      padding: 34px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
    }

    h1 {
      margin: 0;
      font-size: clamp(2rem, 6vw, 3.4rem);
      letter-spacing: 0;
    }

    h2 {
      margin: 4px 0 0;
      color: var(--muted);
      font-size: clamp(1.2rem, 3vw, 1.6rem);
      font-weight: 650;
      letter-spacing: 0;
    }

    p {
      margin: 20px 0 0;
      color: var(--muted);
      font-size: 1.05rem;
    }

    a.button {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-height: 52px;
      margin-top: 26px;
      border-radius: 6px;
      background: var(--accent);
      color: #ffffff;
      padding: 0 22px;
      font-size: 1.08rem;
      font-weight: 700;
      text-decoration: none;
    }

    dl {
      display: grid;
      grid-template-columns: minmax(150px, 0.5fr) minmax(0, 1fr);
      gap: 10px 16px;
      margin: 28px 0 0;
      padding-top: 22px;
      border-top: 1px solid var(--line);
    }

    dt {
      color: var(--muted);
      font-weight: 700;
    }

    dd {
      min-width: 0;
      margin: 0;
      overflow-wrap: anywhere;
      font-weight: 650;
    }

    @media (prefers-color-scheme: dark) {
      :root {
        --bg: #111719;
        --panel: #182124;
        --text: #edf2f3;
        --muted: #a7b4b9;
        --line: #2c3a3f;
        --accent: #4f9fc4;
      }
    }

    @media (max-width: 560px) {
      main {
        width: min(100% - 20px, 760px);
        padding: 24px;
      }

      a.button {
        width: 100%;
      }

      dl {
        grid-template-columns: 1fr;
      }
    }
  </style>
</head>
<body>
  <main>
    <h1>kiosk-client</h1>
    <h2>Willkommen</h2>

    <p>Der kiosk-client wurde erfolgreich installiert.</p>
    <p>Es wurde noch keine Zielseite konfiguriert.</p>
    <p>Bitte öffnen Sie die lokale Client-Administration, um die Zieladresse des Kiosks festzulegen.</p>

    <a class="button" href="/">Client konfigurieren</a>

    <dl>
      <dt>Lokale Administration</dt>
      <dd>http://localhost:8080/</dd>
      <dt>Lokale Administration</dt>
      <dd id="ip-admin">http://-:8080/</dd>
    </dl>
  </main>

  <script>
    fetch("/api/status", { headers: { "Accept": "application/json" } })
      .then((response) => response.ok ? response.json() : null)
      .then((status) => {
        if (!status || !status.ip) {
          return;
        }

        document.getElementById("ip-admin").textContent = "http://" + status.ip + ":8080/";
      })
      .catch(() => {});
  </script>
</body>
</html>`
