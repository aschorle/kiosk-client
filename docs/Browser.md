# Browser

Chromium ist die einzige sichtbare Anwendung der Appliance.

## Start

`kiosk-appliance.service` startet:

```text
dbus-run-session
-> cage
-> scripts/browser-supervisor.sh
-> scripts/start-browser.sh
-> chromium
```

`scripts/browser-supervisor.sh` bleibt als langlebiger Cage-Child-Prozess aktiv. Es startet `scripts/start-browser.sh`, das `config/client.conf` liest, Chromium ermittelt und den Browser im Kioskmodus startet.

## URL

Wenn `URL` leer oder nicht gueltig konfiguriert ist, verwendet das Startskript die lokale Willkommensseite:

```text
http://localhost:8080/welcome
```

Sobald eine gueltige URL gespeichert wurde, startet Chromium mit dieser Zielseite.

## Chromium-Parameter

- `--kiosk`
- `--incognito`
- `--no-first-run`
- `--disable-session-crashed-bubble`
- `--disable-infobars`
- `--disable-background-networking`
- `--disable-background-timer-throttling`
- `--disable-renderer-backgrounding`
- `--disable-sync`
- `--overscroll-history-navigation=0`

## Steuerung

Reload und Neustart werden per Signal an `scripts/browser-supervisor.sh` angefordert:

- `SIGUSR1`: Reload
- `SIGUSR2`: Neustart

Der Supervisor beendet nur sein Chromium-Child und startet es anschliessend neu. Cage bleibt dabei aktiv.
