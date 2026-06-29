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

Installer Architektur

Das Installer-Framework ist modular aufgebaut. In Phase 2 werden ausschließlich Struktur, Zuständigkeiten und Funktionsrahmen vorbereitet. Es werden noch keine produktiven Installationsbefehle ausgeführt.

- `installer/install.sh`: Einstiegspunkt des Installers. Koordiniert später Parameter, Zielplattform, gemeinsame Schritte, Verifikation und Cleanup.
- `installer/install-common.sh`: Gemeinsame Hilfsfunktionen für Logging, Prüfungen, Konfiguration und wiederverwendbare Installer-Logik.
- `installer/install-radxa.sh`: Hardware-spezifischer Ablauf für Radxa Rock 4C+.
- `installer/install-rpi.sh`: Hardware-spezifischer Ablauf für Raspberry Pi 4.
- `installer/packages.sh`: Struktur für Paketquellen, Basispakete und Kiosk-spezifische Pakete.
- `installer/browser.sh`: Struktur für Chromium-Profil, Kiosk-URL und Browser-Startparameter.
- `installer/wayland.sh`: Struktur für Wayland-Umgebung und Cage-Sitzung.
- `installer/systemd.sh`: Struktur für Installation, Aktivierung und Prüfung von systemd-Units.
- `installer/network.sh`: Struktur für NetworkManager und optionale WLAN-Konfiguration.
- `installer/verify.sh`: Struktur für Betriebssystem-, Komponenten- und Abschlussprüfungen.
- `installer/cleanup.sh`: Struktur für temporäre Dateien, optionale Cache-Bereinigung und Abschlussmeldungen.

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
