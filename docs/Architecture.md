Architecture
============

Überblick

Der Kiosk-Client ist bewusst minimalistisch: Er kennt nur eine Ziel-URL, die ihm sagt, welche Webseite im Vollbild angezeigt wird. Alle inhaltlichen Entscheidungen (Playlist, Timing, Medienspeicherung) werden vom Kiosk-Server getroffen.

Kernkomponenten

- Konfiguration: `config/client.conf`
  - URL: Die einzige Anzeige-URL
  - DEVICE_ID: Identifikation des Geräts
  - PORT: Lokaler Web-UI-Port
  - VERSION: Client-Version

- Browser-Stack
  - Kein Desktop-Environment
  - Wayland als Display-Server
  - Cage als Wayland-Composer / Kiosk-Window-Manager
  - Chromium im Kioskmodus (Wayland-Backend)

- Systemintegration
  - systemd-Units (geplant): `kiosk.service`, `kiosk-agent.service`, `kiosk-watchdog.service`
  - NetworkManager für WLAN-Management
  - SSH für Fernwartung

Minimaler Laufzeitablauf

1. systemd startet `kiosk.service` nach Boot.
2. Agent liest `config/client.conf` und stellt lokale Web-UI bereit (http://CLIENT-IP:PORT).
3. Browser (Chromium) wird im Vollbild geladen und zeigt die in der Konfiguration definierte URL.
4. Agent bietet API-Endpunkte (siehe docs/API.md) zum Abfragen und Anpassen der Konfiguration sowie zum Neustarten/Reloaden des Browsers.

Konfigurationsbeispiel

```
URL=https://display.example.local
DEVICE_ID=display01
PORT=8080
VERSION=0.1.0
```

Sicherheit

- SSH-Zugang sollte per Schlüssel abgesichert werden.
- Kommunikation zwischen Client und Server läuft über HTTPS (Server-seitig).

Erweiterbarkeit

Hardware-spezifische Anpassungen werden in den Installer-Skripten abgelegt. Systemd-Units sowie Agent- und Watchdog-Logik werden später als separate Komponenten implementiert.
