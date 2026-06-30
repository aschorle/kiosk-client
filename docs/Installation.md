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
6. `kiosk-agent` im Entwicklungsmodus bauen oder im Release-Modus pruefen
7. systemd user services installieren
8. tty1 Autologin aktivieren

## Pakete

Der Installer installiert nur die Appliance-Pakete:

- `ca-certificates`
- `chromium`
- `cage`
- `dbus`
- `dbus-user-session`

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

Development Mode:

- Wenn `go` vorhanden ist, baut der Installer `kiosk-agent` immer neu.
- Buildfehler brechen die Installation ab.

Release Mode:

- Wenn `go` fehlt, muss ein ausfuehrbares `kiosk-agent`-Binary vorhanden sein.
- Fehlt das Binary, bricht der Installer mit klarer Meldung ab.

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
