Copilot Instructions for kiosk-client
=====================================

Ziel

Helfen beim Aufbau der Projektstruktur, bei Dokumentation und beim Erzeugen von Platzhalterdateien. Kein Produktivcode ohne ausdrückliche Anweisung.

Regeln

- Änderungen an produktivem Bash-/Python-/systemd-Code nur nach ausdrücklicher Freigabe.
- Beim Erzeugen von Dateien Klartext in Deutsch erzeugen (Projekt-Dokumentation auf Deutsch).
- Installer- und Hardware-spezifische Skripte nur als Platzhalter anlegen.
- Tests/CI: keine CI-Skripte anlegen ohne Rücksprache.

Konventionen

- Dateien in `docs/` enthalten architektur- und installationsrelevante Informationen.
- Konfigurationen in `config/` (Beispiel-Dateien mit `.example` suffix).
- Systemd-Unit-Templates in `systemd/` als Platzhalter.
