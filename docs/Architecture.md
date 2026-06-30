# Architecture

Der kiosk-client ist eine lokale Appliance fuer ein einzelnes Geraet.

## Komponenten

- Agent: lokale API und Weboberflaeche
- Browser: lesende Laufzeitinformationen und systemd-Steuerung der Appliance-Runtime
- Config: `config/client.conf`
- Installer: Appliance-Installation
- Runtime: systemd user services, Cage, Chromium
- Status: System-, Browser- und Metrikdaten
- Web: lokales Dashboard und Welcome-Seite

## Grenzen

Es gibt keine Cloud, kein zentrales Management und keine Mehrgeraeteverwaltung.

Neue Funktionen muessen zur lokalen Minimal-Appliance passen.
