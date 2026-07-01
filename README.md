# kiosk-client

Lokale Kiosk-Appliance fuer genau ein Geraet.

Der kiosk-client startet nach dem Boot eine lokale Administrationsoberflaeche und Chromium im Kioskmodus. Chromium zeigt ausschliesslich die konfigurierte URL. Es gibt keine Cloud, kein zentrales Management und keine Mehrgeraeteverwaltung.

## Zielplattform

- Debian oder Armbian Minimal
- systemd
- getty Autologin auf tty1
- systemd user services
- dbus-run-session
- Cage
- Chromium

## Runtime

```text
Boot
-> systemd
-> getty Autologin
-> systemd --user default.target
-> kiosk-agent.service
-> kiosk-appliance.service
-> dbus-run-session
-> scripts/start-cage.sh
-> cage
-> scripts/start-browser.sh
-> Chromium
-> konfigurierte URL
```

## Installation

```bash
sudo KIOSK_USER=rock ./installer/install.sh
```

Der Installer erkennt Debian-/Armbian-Systeme ueber `/etc/os-release` und akzeptiert Bookworm sowie Trixie. Board-spezifische Einstiegspunkte delegieren auf das gemeinsame Appliance-Profil.

Installierte Pakete:

- `ca-certificates`
- `chromium`
- `cage`
- `dbus`
- `dbus-user-session`

## Konfiguration

Die lokale Konfiguration liegt in:

```text
config/client.conf
```

Wichtige Werte:

- `URL`: Zielseite des Kiosks
- `DEVICE_ID`: lokale Geraetekennung
- `BROWSER`: Chromium-Binary, standardmaessig `chromium`
- `AUTH_TOKEN`: optionaler Schreibschutz fuer lokale API-Aufrufe

`AUTH_TOKEN` wird nicht ueber die Weboberflaeche oder JSON-Konfiguration ausgegeben.

## Lokale Administration

Der Agent stellt die lokale Oberflaeche auf Port `8080` bereit:

```text
http://localhost:8080/
```

Verwendete REST-Endpunkte:

- `GET /api/status`
- `GET /api/info`
- `GET /api/config`
- `PUT /api/config`
- `GET /api/health`
- `GET /api/metrics`
- `POST /api/browser/reload`
- `POST /api/browser/restart`

`PUT /api/config` speichert nur die Konfiguration. Browseraktionen werden getrennt ueber die Browser-Endpunkte ausgefuehrt.

## Browsersteuerung

Browseraktionen steuern ausschliesslich die Appliance-Runtime:

```bash
systemctl --user restart kiosk-appliance.service
```

`POST /api/browser/reload` und `POST /api/browser/restart` starten die Appliance-Runtime neu. Dadurch startet Cage Chromium erneut und liest die aktuelle Konfiguration.

## First Boot

Solange keine gueltige Ziel-URL konfiguriert ist, oeffnet Chromium die lokale Willkommensseite:

```text
http://localhost:8080/welcome
```

Nach dem Speichern einer gueltigen URL startet die Appliance-Runtime mit dieser Zielseite.

## Version

Aktuelle Version: `0.12.5`
