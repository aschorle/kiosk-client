# Agent

Der `kiosk-agent` ist der lokale Prozess fuer Administration, API, Status und Metriken.

## Aufgaben

- `config/client.conf` laden und schreiben
- lokale Weboberflaeche ausliefern
- REST-Endpunkte unter `/api/...` bereitstellen
- Status-, Health- und Metrikdaten sammeln
- Browserzustand lesen
- Appliance-Runtime ueber `kiosk-appliance.service` neu starten

Der Agent verwaltet genau dieses eine lokale Geraet.
