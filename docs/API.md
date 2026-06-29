API - Lokale Weboberfläche (Platzhalter)
=========================================

Der Kiosk-Client stellt eine kleine lokale HTTP-API bereit, um Status und Konfiguration anzuzeigen bzw. zu ändern. Diese API ist in späteren Releases zu implementieren. Endpunkte (Entwurf):

GET /api/status
- Rückgabe: JSON mit System- und Browserstatus

GET /api/config
- Rückgabe: JSON mit aktueller Konfiguration (URL, DEVICE_ID, PORT, VERSION)

POST /api/config
- Body: JSON mit neuen Konfigurationswerten (z.B. `{ "URL": "https://..." }`)
- Wirkung: Konfiguration speichern und optional Browser neu laden

POST /api/reload
- Wirkung: Browser-Reload (F5 / navigate reload)

POST /api/restart-browser
- Wirkung: Browser-Prozess neu starten

POST /api/reboot
- Wirkung: Gerät neu starten (systemd reboot)

Sicherheits-Hinweis
- API-Endpunkte sollen nur lokal erreichbar sein (Bind auf localhost oder Firewall-Regeln)
- Authentifizierung/Autorisierung ist später zu spezifizieren (z.B. Token oder SSH-only)

Weitere Hinweise
- Der lokale Web-UI-Port wird über `config/client.conf` festgelegt (Standard: 8080)
- API-Versionierung sollte von Anfang an geplant werden (z.B. `/api/v1/...`)
