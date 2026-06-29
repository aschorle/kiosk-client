Installation (Dokumentation)
===========================

Zielplattform

- Debian Bookworm (Minimal ISO)
- Primär: Radxa Rock 4C+
- Sekundär: Raspberry Pi 4

Hauptprinzipien

- Kein Desktop-Environment
- Verwendung von Wayland + Cage
- Chromium im Kioskmodus (Wayland)
- Systemd zur Steuerung und Überwachung
- NetworkManager für WLAN-Verwaltung

Vorbereitungen

- Minimales Debian auf Zielgerät installieren (Bookworm)
- Basispakete: `sudo apt update` & `sudo apt install --no-install-recommends` (siehe Installer-Skripte)
- Netzwerk-Zugang sicherstellen (temporär Ethernet oder vorkonfigurierte WLAN-Profile)
- SSH-Server installieren und mit Schlüsseln absichern

Installer

Alle install-/setup-spezifischen Schritte werden später als Shell-Skripte unter `installer/` abgelegt. Ziel ist es, für jede Hardware (Radxa, Raspberry Pi) ein angepasstes Installationsskript zu haben, das:

- notwendige Pakete installiert
- Kernel/Firmware-Anpassungen vornimmt (falls nötig)
- systemd-Units installiert und aktiviert
- NetworkManager konfiguriert
- Chromium und Cage für Wayland installiert/configured

Beispielhafte manuelle Schritte (nicht als Produktivskript ausgeführt)

- Paketinstallation (als Hinweis):

```bash
apt update
apt install --no-install-recommends network-manager openssh-server chromium-wayland cage
```

- systemd-Units aktivieren (Platzhalter):

```bash
systemctl enable kiosk.service
systemctl enable kiosk-agent.service
```

WLAN-Konfiguration

NetworkManager wird verwendet, damit Geräte leicht per CLI oder nmcli konfiguriert werden können. Installer-Skripte sollten Profile zur Verfügung stellen und optional einfache interaktive Setup-Hilfen.

Hinweis

Dies ist Dokumentation und keine ausführbaren Installationsskripte. Konkrete, getestete Installationsskripte werden in `installer/` später ergänzt.
