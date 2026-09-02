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

Im expliziten Management-only-Profil nutzt der Agent statt des Supervisors den
fest verdrahteten Systemd-Controller fuer `kiosk.service`. Er meldet dessen
MainPID nur als laufenden Browser, wenn dieser ein Chromium-/Chrome-Prozess im
Kioskmodus ist. Der Agent kann dort nur diesen einzelnen Dienst neu starten;
Reboot und automatische Browserrestarts bleiben deaktiviert.
