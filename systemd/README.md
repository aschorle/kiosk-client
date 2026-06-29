Systemd unit templates (Platzhalter)
====================================

Geplante Services:

- kiosk.service
- kiosk-agent.service
- kiosk-watchdog.service

Hinweis: Die konkreten Unit-Dateien werden später mit genauen ExecStart-Pfaden und Abhängigkeiten ergänzt. Units sollten in `/etc/systemd/system/` installiert und mit `systemctl daemon-reload` aktiviert werden.
