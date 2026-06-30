# API

Der `kiosk-agent` stellt eine lokale HTTP-API auf Port `8080` bereit.

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

Schreibende Endpunkte akzeptieren optional `Authorization: Bearer <AUTH_TOKEN>`.
Wenn `AUTH_TOKEN` leer ist, sind lokale Schreibzugriffe ohne Token erlaubt.

## Konfiguration

`PUT /api/config` schreibt `config/client.conf` und antwortet bei erfolgreichem Speichern mit HTTP 200. Der Browser-Neustart ist davon getrennt und wird ausschliesslich ueber die Browser-Endpunkte ausgefuehrt.

`AUTH_TOKEN` wird nicht ueber `GET /api/config` ausgegeben.

## Browsersteuerung

Reload und Neustart starten die Appliance-Runtime neu:

```text
kiosk-appliance.service
```

Dadurch startet Cage Chromium erneut und liest die aktuelle Konfiguration.
