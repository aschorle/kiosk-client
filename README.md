kiosk-client
=============

Kurzbeschreibung

Ein schlanker Linux-Kiosk-Client für Radxa Rock 4C+ und Raspberry Pi 4. Der Client startet nach dem Boot automatisch eine einzige Webseite im Vollbildmodus (Kiosk). Die Webseite wird von einem separaten Kiosk-Server ausgeliefert; alle Playlist- und Steuerungslogiken liegen auf dem Server.

Ziele

- Schlank, stabil, wartungsarm
- Debian Bookworm minimal (kein Desktop)
- Wayland + Cage
- Chromium im Kioskmodus
- systemd-Services für Start/Überwachung
- Netzwerk über NetworkManager (WLAN-Betrieb)
- SSH für Wartung

Projektstruktur (Auszug)

- .github/copilot-instructions.md
- .vscode/
- docs/
- browser/
- config/client.conf.example
- installer/
- scripts/
- systemd/
- agent/
  - cmd/kiosk-agent/
  - internal/config/
  - internal/browser/
  - internal/status/
  - internal/web/
- web/
- assets/
- go.mod
- README.md
- CHANGELOG.md
- LICENSE
- .gitignore

Erste Schritte

Dies ist ein Dokumentations- und Platzhalter-Repository. Keine Produktionsskripte sind in dieser Phase enthalten. Siehe docs/ für Details zu Architektur, Installation und Hardware.

Projektphasen

Phase 1

- Projektstruktur
- Dokumentation

Phase 2

- Installer Framework

Phase 3

- Browser
- Wayland
- Cage

Phase 4

- systemd

Phase 5

- Kiosk Agent

Phase 6

- Weboberfläche

Phase 7

- Watchdog

Phase 8

- Raspberry Pi Support

Roadmap

Version 0.3

- Runtime X11

Version 0.3.1

- Runtime Hardening

Version 0.4.0

- Runtime Wayland/Cage

Version 0.4.1

- Browser Manager Monitoring

Browser Manager

Der `kiosk-agent` übernimmt ab Version 0.4.1 die lesende Überwachung des Chromium-Browsers. Er ermittelt, ob der Browser läuft, welche PID verwendet wird, welcher Executable-Pfad aktiv ist, welche Version gefunden wird und mit welcher Kommandozeile Chromium gestartet wurde.

Die Browser-Version wird zuerst über bekannte Debian-Pakete wie `chromium` und `chromium-x11` ermittelt. Falls diese Paketinformationen nicht verfügbar sind, fragt der Agent distributionsunabhängig das tatsächlich verwendete Chromium-Binary mit `--version` ab.

Der Agent startet, stoppt und beendet Chromium in dieser Phase nicht. Eine automatische Wiederherstellung oder ein Neustart des Browsers ist erst für Version 0.5 vorgesehen.

Agent Runtime

Ab Version 0.4.2 wird der `kiosk-agent` als systemd User Service betrieben. Die Unit `kiosk-agent.service` läuft im Kontext der grafischen Benutzersitzung, verwendet `%h/kiosk-client` als Arbeitsverzeichnis und startet das Binary `%h/kiosk-client/kiosk-agent`.

Der Installer legt die Unit unter `~/.config/systemd/user/` ab, aktiviert sie und startet sie, sobald `graphical-session.target` aktiv ist. Der Service verwendet `Restart=always` und `RestartSec=3`, damit der lokale Agent nach einem Absturz automatisch wieder verfügbar wird.

Diagnostics

Ab Version 0.4.3 liefert `/api/status` zusätzliche Systemdiagnosen. Dazu gehören Uptime, Kernel-Version, Debian-Version, Architektur, CPU-Modell, Speicher, freier Speicherplatz und Load Average.

Alle Werte werden ausschließlich lesend über `/proc`, `/sys`, `uname` beziehungsweise Go-Standardbibliothek ermittelt. Der Agent benötigt dafür keine Root-Rechte, startet keine externen Programme und verändert keine Systemzustände.

Configuration API

Ab Version 0.4.4 stellt der `kiosk-agent` die beim Start geladene Client-Konfiguration über `GET /api/config` bereit. Die Antwort enthält `url`, `device_id` und `browser` als JSON.

Die API ist ausschließlich lesend. `config/client.conf` wird nicht verändert und für diese Route nicht erneut geöffnet. Wenn die Konfiguration beim Start nicht geladen oder nicht gültig war, antwortet der Agent mit HTTP 500 und einem JSON-Fehlerobjekt.

Information API

Ab Version 0.4.5 stellt `GET /api/info` allgemeine Agent-, Build-, Betriebssystem- und Board-Informationen als JSON bereit. Enthalten sind `agent_version`, `go_version`, `hostname`, `architecture`, `kernel`, `build_time`, `git_commit`, `board`, `os_name` und `os_version`.

Die Werte werden ausschließlich lesend über Go-Standardbibliothek und vorhandene Statusfunktionen ermittelt. Wenn `build_time` oder `git_commit` beim Build nicht gesetzt wurden, gibt der Agent jeweils `unknown` zurück.

Runtime Control API

Ab Version 0.4.6 stellt der `kiosk-agent` einfache Steuer-Endpunkte für den Browser-Service bereit. `POST /api/browser/restart` führt `systemctl --user restart kiosk-browser.service` aus und `POST /api/browser/reload` führt nur dann `systemctl --user reload kiosk-browser.service` aus, wenn systemd für die Unit Reload-Unterstützung meldet.

Erfolgreiche Aufrufe antworten mit `{"status":"ok"}`. Fehler werden als JSON mit `{"error":"..."}` zurückgegeben. Wenn Reload für `kiosk-browser.service` nicht unterstützt wird, antwortet die API mit HTTP 501 und `{"error":"reload not supported"}`.

Die Runtime Control API nutzt ausschließlich den systemd User Manager. Es gibt keine direkte Prozesssteuerung, keine Signale, keine Shell-Skripte und keine `kill()`-Aufrufe.

Browser Watchdog

Ab Version 0.5.0 startet der `kiosk-agent` einen Hintergrund-Worker, der alle 30 Sekunden prüft, ob Chromium läuft. Wenn der Browser nicht sichtbar ist, löst der Agent über den systemd User Manager einen Restart von `kiosk-browser.service` aus.

Jeder erfolgreiche Watchdog-Restart erhöht `browser_restart_count` und setzt `browser_last_restart` auf die aktuelle UTC-Zeit. Beide Werte werden über `/api/status` ausgeliefert. Der Worker wird über einen Context gesteuert und beendet sich beim Programmende sauber.

Neue Runtime Architektur

Boot

↓

Autologin

↓

kiosk-runtime.service

↓

Cage

↓

Chromium
