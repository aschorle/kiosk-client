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

Go Runtime

Der kiosk-client verwendet ab Version 0.4.0 einen Go-basierten `kiosk-agent`. Damit der Agent direkt auf dem Zielsystem gebaut werden kann, gehört `golang-go` zu den gemeinsamen Basispaketen.

Nach:

```bash
sudo ./installer/install.sh
```

steht der Befehl `go` automatisch auf dem System zur Verfügung. Die Installation erfolgt idempotent über die bestehende Paketliste `COMMON_PACKAGES`; erneute Installer-Läufe installieren Go nicht doppelt, sondern halten das Paket lediglich vorhanden.

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

Appliance Runtime

Ab Version 0.8.0 wird der produktive Browserstart ueber Cage ausgefuehrt. GNOME und GDM bleiben installiert und werden nicht entfernt. GDM stellt weiterhin den Autologin bereit; danach startet systemd im Benutzerkontext die Appliance Runtime.

Der produktive Ablauf ist:

```text
Autologin
-> systemd user service
-> kiosk-runtime.service
-> scripts/start-cage.sh
-> cage
-> scripts/start-browser.sh
-> chromium
-> URL aus config/client.conf
```

Installation und Aktivierung erfolgen weiterhin ueber den Hauptinstaller:

```bash
sudo ./installer/install.sh
```

`installer/cage.sh` installiert das Paket `cage` idempotent und prueft die Runtime mit `command -v cage` sowie `cage -v`. `installer/systemd.sh` installiert die User-Services nach `~/.config/systemd/user/`. Dabei werden `kiosk-agent.service` und `kiosk-runtime.service` aktiviert. `kiosk-browser.service` wird nur noch als Legacy/Fallback-Datei installiert, aber nicht mehr aktiviert.

Wenn `kiosk-browser.service` aus einer frueheren Version noch enabled ist, deaktiviert der Installer den Service. Wenn er noch laeuft, wird er gestoppt. `kiosk-runtime.service` wird anschliessend nur dann direkt gestartet, wenn `graphical-session.target` in der User-Session aktiv ist. Andernfalls startet die Runtime beim naechsten grafischen Login oder Boot.

Fallback auf den alten Browser-Service:

```bash
systemctl --user disable kiosk-runtime.service
systemctl --user stop kiosk-runtime.service
systemctl --user enable kiosk-browser.service
systemctl --user restart kiosk-browser.service
```

Zurueck zur Appliance Runtime:

```bash
systemctl --user disable kiosk-browser.service
systemctl --user stop kiosk-browser.service
systemctl --user enable kiosk-runtime.service
systemctl --user restart kiosk-runtime.service
```

Debugging-Befehle:

```bash
systemctl --user status kiosk-runtime.service
systemctl --user status kiosk-browser.service
journalctl --user -u kiosk-runtime.service -f
journalctl --user -u kiosk-browser.service -f
command -v cage
cage -v
```

Native Kiosk Session

Ab Version 0.8.1 meldet der Display Manager den Kiosk-Benutzer direkt in eine eigene Session `kiosk` an. Diese Session startet nur `scripts/start-cage.sh`. Damit laeuft der kiosk-client nicht mehr innerhalb einer KDE-, Plasma-, GNOME- oder X11-Desktop-Sitzung.

`kiosk-runtime.service` bleibt installiert, wird im nativen Sessionbetrieb aber nicht mehr automatisch aktiviert. So wird verhindert, dass Cage einmal durch den Display Manager und ein zweites Mal durch systemd gestartet wird.

Unterstuetzte Display Manager:

- GDM/GDM3
- SDDM
- LightDM

Der Installer erkennt den installierten Display Manager automatisch ueber `/etc/X11/default-display-manager`, bekannte Konfigurationsverzeichnisse und installierte Display-Manager-Binaries. Die Session-Datei wird als Wayland-Session installiert:

```text
/usr/share/wayland-sessions/kiosk.desktop
```

Die Session startet:

```text
<repo>/scripts/start-cage.sh
```

`scripts/start-cage.sh` ersetzt den laufenden Session-Prozess direkt mit:

```text
exec cage -- scripts/start-browser.sh
```

Es bleiben keine Shell-Wrapper oder Hintergrundprozesse zwischen Display Manager und Cage bestehen.

Autologin-Session je Display Manager:

- GDM/GDM3: `installer/session.sh` setzt den Autologin-Benutzer in `/etc/gdm3/daemon.conf` und die Session `kiosk` in `/var/lib/AccountsService/users/<user>`.
- SDDM: `installer/session.sh` schreibt `/etc/sddm.conf.d/kiosk-client.conf` mit `User=<user>` und `Session=kiosk.desktop`.
- LightDM: `installer/session.sh` schreibt `/etc/lightdm/lightdm.conf.d/50-kiosk-client.conf` mit `autologin-user=<user>` und `autologin-session=kiosk`.

GDM merkt sich die bevorzugte Sitzung eines Benutzers ueber AccountsService. Wenn dort noch Plasma oder GNOME hinterlegt ist, kann der Autologin trotz vorhandener `kiosk.desktop` wieder `startplasma-x11`, `plasmashell` und `kwin_x11` starten. Deshalb muss fuer GDM in `/var/lib/AccountsService/users/<user>` im Abschnitt `[User]` explizit stehen:

```ini
[User]
Session=kiosk
```

Andere Inhalte in dieser Datei bleiben erhalten. Der Installer setzt die Datei auf `root:root` und `0644`. Falls `accounts-daemon.service` vorhanden ist, wird er nach der Aenderung neu gestartet; fehlt der Dienst, laeuft die Installation weiter.

Ab Version 0.8.3 wird `kiosk` nur als Wayland-Session registriert. Eine alte vom kiosk-client erzeugte `/usr/share/xsessions/kiosk.desktop` wird entfernt, weil GDM sonst den X11-Sessionpfad verwenden kann. In diesem Fall kann `loginctl` weiter `Type=x11` anzeigen und der Kiosk wirkt nicht wie eine exklusive Cage-Session.

Der Installationsablauf bleibt:

```bash
sudo ./installer/install.sh
```

Fallback auf Plasma, GNOME oder eine andere vorhandene Desktop-Session:

1. Autologin im Display Manager deaktivieren oder auf die gewuenschte Desktop-Session umstellen.
2. Im grafischen Login-Bildschirm eine vorhandene Session wie Plasma oder GNOME auswaehlen.
3. Die Datei `kiosk.desktop` kann installiert bleiben; sie ist nur eine zusaetzliche Session-Auswahl.

Debugging-Befehle:

```bash
cat /etc/X11/default-display-manager
ls -l /usr/share/wayland-sessions/kiosk.desktop
cat /usr/share/wayland-sessions/kiosk.desktop
test ! -e /usr/share/xsessions/kiosk.desktop
cat /etc/gdm3/daemon.conf
cat /var/lib/AccountsService/users/$USER
grep '^Session=' /var/lib/AccountsService/users/$USER
cat /etc/sddm.conf.d/kiosk-client.conf
cat /etc/lightdm/lightdm.conf.d/50-kiosk-client.conf
journalctl -b
```

Native Session Debugging:

```bash
loginctl
loginctl show-session <SESSION_ID> -p Type -p Name -p Desktop -p State
pgrep -a cage
pgrep -a start-cage
pgrep -a startplasma-x11
pgrep -a plasmashell
pgrep -a kwin_x11
```

Erwartet wird eine `kiosk`-Session ohne `startplasma-x11`, `plasmashell` oder `kwin_x11`. `start-cage.sh` darf nach dem Start nicht als dauerhafter Wrapper-Prozess bestehen bleiben, weil es sich per `exec` durch Cage ersetzt.

Phase 3

Wayland/Cage wurde vorbereitet, ist jedoch bis zur Umstellung auf ein aktuelleres Chromium deaktiviert. Die aktuelle Radxa-Version von Chromium unterstützt `--ozone-platform=wayland` nicht zuverlässig. Deshalb läuft die Runtime in Version 0.3 weiterhin über X11, während Cage installiert bleibt und für Version 0.4 vorbereitet ist.

Kiosk Runtime

Version 0.3.1 härtet die Kiosk-Laufzeit gegen Unterbrechungen durch Energiespar- und Sperrfunktionen. Der Bildschirm bleibt dauerhaft aktiv, es gibt keinen Lock Screen, keinen Bildschirmschoner, kein automatisches Abdunkeln, keine automatische Displayabschaltung und keinen automatischen Suspend oder Hibernate.

Die Konfiguration erfolgt über `installer/power.sh`. Für GNOME/GDM3 werden offizielle `gsettings` verwendet, um Sperrbildschirm, Bildschirmschoner, Abdunkeln und automatische GNOME-Schlafzustände zu deaktivieren. Systemweite Schlaf- und Idle-Aktionen werden über systemd-Konfigurationsdateien unter `/etc/systemd/logind.conf.d/` und `/etc/systemd/sleep.conf.d/` deaktiviert.

Es werden keine `xset`-Workarounds und keine Shell-Hacks verwendet.

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
