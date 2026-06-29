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
