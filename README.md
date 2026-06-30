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
- web/
- assets/
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

Version 0.4

- Runtime Wayland/Cage
