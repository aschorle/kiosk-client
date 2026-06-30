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

Vorabprüfungen

Die Vorabprüfungen sind reine Leseoperationen. Sie bereiten spätere Installer-Phasen vor, verändern das System aber nicht.

- Debian Version: prüft, ob Debian 12 Bookworm als unterstütztes Basissystem verwendet wird.
- Root-Rechte: prüft, ob der Installer mit administrativen Rechten ausgeführt wird.
- Netzwerkverbindung: prüft grundlegende Namensauflösung, ohne Netzwerkprofile zu verändern.
- Freier Speicherplatz: prüft den verfügbaren Speicherplatz auf dem Root-Dateisystem gegen einen konservativen Mindestwert.
- CPU Architektur: prüft, ob die Architektur zu den unterstützten ARM-Zielplattformen passt.
- Unterstütztes Board: erkennt Radxa Rock 4C+ oder Raspberry Pi 4 über Geräteinformationen, ohne Hardware-Konfiguration zu ändern.

Grundinstallation auf einer Radxa Rock 4C+

Die Grundinstallation für Radxa Rock 4C+ wird über `installer/install-radxa.sh` gestartet. Das Skript ist hardware-spezifisch und verwendet die gemeinsamen Funktionen aus `install-common.sh`, die Paketlisten aus `packages.sh` und die Vorabprüfungen aus `verify.sh`.

Der Ablauf ist bewusst klein gehalten:

- Root-Rechte prüfen.
- Vorabprüfungen für Debian-Version, Netzwerk, Speicherplatz, CPU-Architektur und unterstütztes Board ausführen.
- Board erkennen und sicherstellen, dass das Skript auf einer Radxa Rock 4C+ läuft.
- Begrüßung und Zusammenfassung der Grundinstallation ausgeben.
- Paketquellen mit `apt update` aktualisieren.
- Vorhandene Pakete mit `apt full-upgrade -y` aktualisieren.
- Ausschließlich die gemeinsamen Basispakete aus `COMMON_PACKAGES` installieren.
- Erfolgsmeldung ausgeben.

In diesem Schritt werden noch kein Chromium, kein Cage, keine Wayland-Konfiguration, keine systemd-Services, kein Kiosk-Agent und kein Webinterface installiert oder eingerichtet.

Browser

Chromium wird als Browser-Komponente verwendet, weil es unter Debian Bookworm verfügbar ist, moderne Webstandards unterstützt und später zuverlässig im Kioskmodus mit Wayland und Cage betrieben werden kann. Die Browser-Komponente liegt in `installer/browser.sh` und bleibt von der Board-Grundinstallation getrennt.

Das Skript installiert ausschließlich das Paket `chromium` über `apt`, prüft danach den tatsächlichen Programmpfad per `command -v` und liest die installierte Version über den Chromium-Aufruf mit `--version` aus. Diese geprüfte Laufzeitversion ist maßgeblich, nicht nur der Paketname.

Der Kioskmodus wird in diesem Schritt noch nicht aktiviert. Browserflags, Policies, Cache-Konfiguration, Autostart, Wayland, Cage, systemd-Integration und URL-Konfiguration folgen in späteren Phasen, damit Installation, Browser-Laufzeit und Kiosk-Verhalten getrennt testbar bleiben.

Automatischer Browserstart

Der automatische Browserstart wird als systemd User Service eingerichtet. Die Unit liegt im Repository unter `systemd/user/kiosk-browser.service` und wird nach `~/.config/systemd/user/kiosk-browser.service` installiert. Sie startet `scripts/start-browser.sh`, wodurch Chromium im Kioskmodus mit der URL aus `config/client.conf` geöffnet wird.

User Service

Chromium muss grundsätzlich innerhalb der grafischen Benutzersitzung laufen. Der Browser benötigt die Sitzungsumgebung des angemeldeten Benutzers und soll nicht als root oder als globaler System-Service gestartet werden. Deshalb verwendet der kiosk-client einen systemd User Service mit `After=graphical-session.target` und `PartOf=graphical-session.target`.

Die Installation erfolgt als Zielbenutzer, nicht mit `sudo`:

```bash
./installer/systemd.sh
```

Das Skript installiert die Unit nach `~/.config/systemd/user/`, führt `systemctl --user daemon-reload` aus, aktiviert den Service und startet ihn direkt neu:

```bash
systemctl --user daemon-reload
systemctl --user enable kiosk-browser.service
systemctl --user restart kiosk-browser.service
```

Der Service verwendet `Restart=always` und `RestartSec=5`. Wenn Chromium abstürzt oder beendet wird, startet systemd den Browser nach fünf Sekunden erneut.

Status prüfen:

```bash
systemctl --user status kiosk-browser.service
```

Deaktivierung:

```bash
systemctl --user disable --now kiosk-browser.service
```

In dieser Phase werden noch kein Cage, keine Wayland-spezifische Konfiguration, kein Watchdog und keine lokale Weboberfläche eingerichtet. Es werden außerdem keine `DISPLAY`-, `XAUTHORITY`- oder sonstigen Sitzungs-Workarounds gesetzt.

GDM3 Autologin

Damit der systemd User Service nach dem Boot eine grafische Benutzersitzung vorfindet, richtet `installer/autologin.sh` automatisches Login für GDM3 ein. Unterstützt wird in dieser Phase ausschließlich GDM3.

Das Skript bearbeitet `/etc/gdm3/daemon.conf` und setzt im Abschnitt `[daemon]`:

```ini
AutomaticLoginEnable=True
AutomaticLogin=<KIOSK_USER>
```

Der Benutzer wird über `KIOSK_USER` bestimmt. Wenn `KIOSK_USER` nicht gesetzt ist, verwendet der Installer `SUDO_USER`. Vor jeder Änderung wird eine Sicherung unter `/etc/gdm3/daemon.conf.bak` angelegt. Die Änderung ist idempotent: Wenn die Werte bereits korrekt gesetzt sind, wird die Datei nicht erneut verändert.

Cage Runtime

Cage wird als schlanker Wayland-Compositor für den kiosk-client vorbereitet. Das Ziel ist eine Laufzeit, in der nach dem Boot keine vollständige Desktop-Umgebung sichtbar ist, sondern nur eine einzelne Anwendung: Chromium im Kioskmodus mit der URL aus `config/client.conf`.

Gegenüber KDE oder einer anderen vollwertigen Desktop-Umgebung hat Cage mehrere Vorteile für einen Kiosk-Client:

- Weniger sichtbare Oberfläche und weniger Ablenkung für den Benutzer.
- Weniger Hintergrunddienste und damit weniger bewegliche Teile.
- Eine klare Single-App-Architektur: genau eine grafische Anwendung steht im Vordergrund.
- Bessere Trennung zwischen Anzeige-Laufzeit und Konfiguration.

Die Zielarchitektur besteht aus drei Runtime-Komponenten:

- Cage: stellt später die minimale Wayland-Sitzung bereit.
- Browser: `scripts/start-browser.sh` startet Chromium im Kioskmodus und ist bereits für Wayland/Ozone vorbereitet.
- Konfiguration: `config/client.conf` enthält die URL und browserbezogene Laufzeitwerte.

In dieser Phase installiert `installer/cage.sh` nur die Cage-Paketbasis. Es wird noch keine vollständige Cage-Sitzung konfiguriert, KDE wird nicht entfernt und der Bootprozess wird noch nicht auf Cage umgeschaltet.

Phase 3

Wayland/Cage wurde vorbereitet, ist jedoch bis zur Umstellung auf ein aktuelleres Chromium deaktiviert. Die aktuelle Radxa-Version von Chromium unterstützt `--ozone-platform=wayland` nicht zuverlässig. Deshalb läuft die Runtime in Version 0.3 weiterhin über X11, während Cage installiert bleibt und für Version 0.4 vorbereitet ist.

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
