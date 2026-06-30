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
