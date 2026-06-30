# Browser

Chromium ist die einzige sichtbare Anwendung der Appliance.

## Start

`kiosk-appliance.service` startet:

```text
dbus-run-session
-> cage
-> scripts/start-browser.sh
-> chromium
```

`scripts/start-browser.sh` liest `config/client.conf`, ermittelt Chromium und startet den Browser im Kioskmodus.

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
- `--disable-gpu`
- `--disable-crash-reporter`
- `--disable-breakpad`
- `--disable-background-networking`
- `--disable-background-timer-throttling`
- `--disable-renderer-backgrounding`
- `--disable-sync`
- `--overscroll-history-navigation=0`

## Steuerung

Reload und Neustart erfolgen ueber:

```text
kiosk-appliance.service
```

Die lokale API startet diese Unit neu. Dadurch startet Cage Chromium erneut und liest die aktuelle Konfiguration.
