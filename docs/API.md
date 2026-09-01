# API

Der `kiosk-agent` stellt eine lokale HTTP-API auf `127.0.0.1:8080` bereit.

## Lesende Endpunkte

- `GET /api/status`
- `GET /api/info`
- `GET /api/config`
- `GET /api/health`
- `GET /api/metrics`

## Schreibende Endpunkte

- `PUT /api/config`
- `POST /api/browser/reload`
- `POST /api/browser/restart`
- `POST /api/system/reboot`

Schreibende Endpunkte akzeptieren optional `Authorization: Bearer <AUTH_TOKEN>`.
Wenn `AUTH_TOKEN` leer ist, sind lokale Schreibzugriffe ohne Token erlaubt.
Da der Agent nur auf Loopback bindet, sind diese Endpunkte nicht über das LAN
erreichbar. Für zentrale Verwaltung muss trotzdem ein `AUTH_TOKEN` gesetzt sein.

## Zentrale Client-Verwaltung

Ist `SERVER_URL` konfiguriert und `DEVICE_ID` sowie `AUTH_TOKEN` nicht leer,
sendet der Agent alle 20 Sekunden einen ausgehenden `POST`
`/api/clients/heartbeat`. Der Bearer-Token wird nur an diese konfigurierte URL
gesendet. Der Payload enthält ausschließlich Kennung, Name, Typ, Agent-Version,
Browserstatus, Health, die beiden Capabilities und Command-Quittungen.

Der Agent führt ausschließlich die Commands `restart_browser` und `reboot` aus;
andere Antworten werden als fehlgeschlagen quittiert. Es werden keine Shell-
Kommandos oder servergelieferten Argumente ausgeführt.

Die zentrale Steuerung ist damit auf das Phase-1-Protokoll beschraenkt: der
Agent bleibt auf `127.0.0.1:8080` gebunden und baut nur die ausgehende
Verbindung zu `SERVER_URL` auf. SSH dient ausschliesslich Installation,
Wartung, Diagnose und Recovery und nicht regulaeren Server-Kommandos.

## Konfiguration

`PUT /api/config` schreibt `config/client.conf` und antwortet bei erfolgreichem Speichern mit HTTP 200. Der Browser-Neustart ist davon getrennt und wird ausschliesslich ueber die Browser-Endpunkte ausgefuehrt.

`AUTH_TOKEN` wird nicht ueber `GET /api/config` ausgegeben.

## Browsersteuerung

Reload und Neustart werden per Signal an den Browser-Supervisor angefordert:

```text
scripts/browser-supervisor.sh
```

Der Supervisor startet Chromium innerhalb der laufenden Cage-Sitzung neu.

## System

`POST /api/system/reboot` loest ueber den lokalen Agent einen sauberen System-Reboot aus.
