# Installation

Der kiosk-client unterstuetzt ausschliesslich die Appliance Edition auf Debian oder Armbian Minimal.

## Voraussetzungen

- Debian/Armbian Bookworm oder Trixie
- systemd
- lokaler Benutzer fuer den Kiosk
- Netzwerkzugang fuer `apt`
- Root-Rechte fuer die Installation

## Start

```bash
sudo KIOSK_USER=rock ./installer/install.sh
```

Wenn `KIOSK_USER` nicht gesetzt ist, verwendet der Installer den Benutzer aus `SUDO_USER`.

## Ablauf

1. Root-Rechte pruefen
2. Architektur pruefen
3. Debian-/Armbian-Version ueber `/etc/os-release` pruefen
4. Board erkennen
5. Appliance-Pakete installieren
6. Go bei Bedarf installieren
7. `kiosk-agent` immer aus den aktuellen Repository-Quellen bauen
8. sudoers-Regel fuer System-Reboot installieren
9. systemd user services installieren
10. tty1 Autologin aktivieren

## Pakete

Der Installer installiert nur die Appliance-Pakete:

- `ca-certificates`
- `chromium`
- `cage`
- `dbus`
- `dbus-user-session`
- `fonts-noto-color-emoji`

## systemd

Installierte User-Units:

```text
~/.config/systemd/user/kiosk-agent.service
~/.config/systemd/user/kiosk-appliance.service
```

Beide Units werden fuer `default.target` aktiviert. Der Startpfad ist:

```text
getty@tty1
-> Autologin
-> systemd --user
-> kiosk-agent.service
-> kiosk-appliance.service
```

## Agent-Binary

Der Installer baut `kiosk-agent` bei jeder erfolgreichen Installation neu aus den aktuellen Repository-Quellen.

- Wenn `go` fehlt, installiert der Installer `golang-go` automatisch.
- Buildfehler brechen die Installation ab.
- Rechte und Besitzer des erzeugten Binaries werden nach dem Build gesetzt.

## Abschluss

Nach erfolgreicher Installation erfolgt kein automatischer Reboot. Der Installer fordert am Ende zu folgendem Befehl auf:

```bash
sudo reboot
```

## Debugging

```bash
systemctl status getty@tty1.service
systemctl --user status kiosk-agent.service
systemctl --user status kiosk-appliance.service
journalctl --user -u kiosk-agent.service -f
journalctl --user -u kiosk-appliance.service -f
pgrep -a cage
pgrep -a chromium
curl http://localhost:8080/api/health
```
