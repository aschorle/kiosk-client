# Agent

Der `kiosk-agent` ist der lokale Prozess fuer Administration, API, Status und Metriken.

## Aufgaben

- `config/client.conf` laden und schreiben
- lokale Weboberflaeche ausliefern
- REST-Endpunkte unter `/api/...` bereitstellen
- Status-, Health- und Metrikdaten sammeln
- Browserzustand lesen
- Browseraktionen per Signal an den Browser-Supervisor anfordern

Der Agent verwaltet genau dieses eine lokale Geraet.
