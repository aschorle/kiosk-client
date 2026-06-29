Roadmap
=======

Kurzfristig (MVP)

- Dokumentation und Platzhalterstruktur (dieses Repository)
- Installer-Skripte für Radxa und Raspberry Pi (Basispakete)
- systemd-Unit-Templates (Platzhalter)
- Lokale Web-UI: Anzeige/Aktualisierung der URL, Reload, Restart

Mittelfristig

- Implementierung des `kiosk-agent` mit oben beschriebener API
- robuster Watchdog-Service zur Überwachung des Browsers
- sichere Remote-Wartungs-Workflows (SSH + Logging)
- optionales Remote-Logging/Monitoring (leichtgewichtig)

Langfristig

- Fernverwaltung/Deploy-Tooling für Massen-Rollouts
- OTA-Update-Strategie für Client-Software
- Feinere Energieoptimierungen und Temperaturmanagement

Prioritäten

1. Stabilität und niedriger Ressourcenverbrauch
2. Klare Trennung von Client- und Server-Logik
3. Einfache, reproduzierbare Installer für die unterstützten Boards
