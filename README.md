# kiosk-client

Lokale Kiosk-Appliance fuer genau ein Geraet.

Der kiosk-client startet nach dem Boot eine lokale Administrationsoberflaeche und Chromium im Kioskmodus. Chromium zeigt ausschliesslich die konfigurierte URL. Phase 1 erlaubt optional eine ausgehende, zentral verwaltete Heartbeat-Verbindung; es gibt keine von aussen erreichbare Remote-API.

## Zielplattform

- Debian oder Armbian Minimal
- systemd
- getty Autologin auf tty1
- systemd user services
- dbus-run-session
- Cage
- Chromium

## Referenzclient und SSH

Der produktive Referenzclient ist `raspi-kiosk` (Raspberry Pi 4 Model B) mit
Armbian 26.5.1 auf Debian trixie und Kernel `6.18.35-current-bcm2711`.

Fuer Installation, Wartung, Diagnose und Recovery wird der lokale SSH-Alias
`raspi-kiosk` mit dem Benutzer `aschorle` verwendet. Er nutzt ausschliesslich
den dedizierten Key `~/.ssh/id_ed25519_raspi_kiosk`. SSH ist nicht der Weg fuer
regulaere zentrale Client-Kommandos.

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
-> scripts/browser-supervisor.sh
-> scripts/start-browser.sh
-> Chromium
-> konfigurierte URL
```

## Installation

```bash
sudo KIOSK_USER=rock ./installer/install.sh
```

Der Installer erkennt Debian-/Armbian-Systeme ueber `/etc/os-release` und akzeptiert Bookworm sowie Trixie. Board-spezifische Einstiegspunkte delegieren auf das gemeinsame Appliance-Profil.

Nach erfolgreicher Installation erfolgt kein automatischer Reboot. Der Installer fordert am Ende ausdruecklich zu `sudo reboot` auf.

Installierte Pakete:

- `ca-certificates`
- `chromium`
- `cage`
- `dbus`
- `dbus-user-session`
- `fonts-noto-color-emoji`

## Konfiguration

Die lokale Konfiguration liegt in:

```text
config/client.conf
```

Wichtige Werte:

- `URL`: Zielseite des Kiosks
- `DEVICE_ID`: lokale Geraetekennung
- `BROWSER`: Chromium-Binary, standardmaessig `chromium`
- `AUTH_TOKEN`: Schreibschutz für lokale API-Aufrufe und Bearer-Token für die zentrale Verwaltung
- `SERVER_URL`: optionale zentrale Server-URL; leer deaktiviert den Heartbeat
- `DEVICE_NAME`: optionaler Anzeigename im zentralen Admin

`AUTH_TOKEN` wird nicht ueber die Weboberflaeche oder JSON-Konfiguration ausgegeben.

`URL` und `SERVER_URL` haben getrennte Aufgaben: `URL` ist die Browser- bzw.
Player-Zieladresse. `SERVER_URL` wird nur fuer die ausgehende zentrale
Verwaltung und Heartbeat-Kommunikation verwendet; sie aendert die Browser-Zieladresse nicht.

## Lokale Administration

Der Agent stellt die lokale Oberflaeche ausschließlich auf `127.0.0.1:8080` bereit:

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
- `POST /api/system/reboot`

`PUT /api/config` speichert nur die Konfiguration. Browseraktionen werden getrennt ueber die Browser-Endpunkte ausgefuehrt.

Der Systemstatus enthaelt die CPU-Temperatur aus `/sys/class/thermal/thermal_zone0/temp`, sofern der Kernel diesen Wert bereitstellt.

`POST /api/system/reboot` loest einen sauberen System-Reboot ueber den lokalen Agent aus. Der Installer richtet dafuer eine eingeschraenkte sudoers-Regel fuer `/usr/bin/systemctl reboot` ein.

## Browsersteuerung

Browseraktionen senden Signale an `scripts/browser-supervisor.sh`. Der Supervisor startet Chromium innerhalb der laufenden Cage-Sitzung neu; Cage bleibt dabei aktiv. `kiosk-appliance.service` bleibt nur fuer den kompletten Runtime-Crash zustaendig.

Unter Wayland/Cage kann der Mauszeiger je nach Plattform sichtbar bleiben. Dies ist eine bekannte Einschränkung der verwendeten Grafikarchitektur und hat keine funktionalen Auswirkungen auf den Appliance-Betrieb.

## First Boot

Solange keine gueltige Ziel-URL konfiguriert ist, oeffnet Chromium die lokale Willkommensseite:

```text
http://localhost:8080/welcome
```

Nach dem Speichern einer gueltigen URL startet die Appliance-Runtime mit dieser Zielseite.

## Linux-Go-Validierung

Vor einem Commit werden die geaenderten Go-Dateien mit `gofmt -w` formatiert.
Die vollstaendige Validierung erfolgt in einer temporaeren Linux-Testkopie ohne
produktive `config/client.conf` und ohne Tokens:

```bash
go test ./...
go vet ./...
```

Anschliessend werden lokal `git diff --check`, `git status --short` und der
Diff geprueft. Die produktive Installation und ihre Dienste bleiben dabei
unveraendert.

## Version

Aktuelle Version: `0.13.6`

## Management-only Client (Phase 1B)

Neben der Raspberry-Appliance unterstuetzt derselbe Agent einen explizit
konfigurierten Management-only-Modus fuer den bestehenden Mini-PC. Er ist kein
Ersatz fuer dessen Display-Stack und erkennt keine Hardware automatisch.

| Profil | Controller | Watchdog | Capabilities | Lokale API |
| --- | --- | --- | --- | --- |
| Appliance (Default) | Browser-Supervisor | `restart` | `restart_browser`, `reboot` | `127.0.0.1:8080` |
| Management-only | festes `kiosk.service` | `observe` | `restart_browser` | `127.0.0.1:18080` |

Management-only verwendet `CLIENT_TYPE=systemd`,
`BROWSER_CONTROLLER=systemd-service`, `BROWSER_SERVICE=kiosk.service`,
`BROWSER_WATCHDOG=observe`, `ENABLE_REBOOT=false`,
`HTTP_ADDR=127.0.0.1:18080` und `CONFIG_WRITABLE=false`. Die vollständige
Vorlage steht in `config/client.conf.management-only.example`.

Der Controller akzeptiert keine aus Remote-Daten stammenden Unit-Namen. Ein
Remote-`restart_browser` ruft ausschliesslich `sudo -n /usr/bin/systemctl
restart kiosk.service` auf und prueft anschliessend den Zustand der festen Unit
und ihren Chromium-Kiosk-MainPID. `reboot` wird in diesem Profil weder gemeldet
noch ausgefuehrt. Der beobachtende Watchdog konkurriert deshalb nicht mit
`Restart=always` von `kiosk.service`.

Der separate Installer `installer/install-management-only.sh` ist nur fuer
einen spaeteren, explizit freigegebenen Ubuntu-24.04-amd64-Pilot vorgesehen.
Er installiert ausschliesslich den Agenten, `/etc/kiosk-agent/client.conf`,
die Agent-Systemunit und eine einzelne sudoers-Regel fuer den festen
Browserrestart. Er beruehrt niemals LightDM, Xorg, LXDE, Openbox, Chromium,
`kiosk.service`, nginx, Port 8080, `/srv/kiosk`, `/srv/kiosk/content`,
`kiosk-playlist.service`, `kiosk-save.service`, `kiosk-command.service` oder
`filebrowser.service`. Die lokale Legacy-Slideshow auf Port 8080 bleibt bis zu
einer separaten Migration produktiv.
