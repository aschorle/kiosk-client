# Appliance Edition

Die Appliance Edition installiert den kiosk-client ohne Display Manager und ohne Desktop Environment. GNOME, KDE, GDM, SDDM oder LightDM werden nicht installiert, konfiguriert oder entfernt.

## Zielarchitektur

```text
Boot
-> systemd getty@tty1
-> Autologin des Kiosk-Benutzers
-> systemd --user default.target
-> kiosk-appliance.service
-> dbus-run-session
-> cage
-> scripts/start-browser.sh
-> Chromium
-> URL aus config/client.conf
```

## Installation

Auf einem minimalen Debian 12 Bookworm:

```bash
sudo KIOSK_USER=rock ./installer/appliance.sh
```

Wenn `KIOSK_USER` nicht gesetzt ist, verwendet der Installer `SUDO_USER`.

## Installierte Pakete

Die Appliance Edition installiert nur:

- `chromium`
- `cage`
- `dbus`

Es werden keine Display-Manager- oder Desktop-Pakete installiert.

## tty Autologin

`installer/tty.sh` schreibt eine systemd-Override-Datei fuer `getty@tty1`:

```text
/etc/systemd/system/getty@tty1.service.d/kiosk-autologin.conf
```

Diese meldet den Kiosk-Benutzer automatisch auf tty1 an. GDM, SDDM und LightDM bleiben unberuehrt.

## Runtime

`installer/runtime.sh` installiert den User Service:

```text
~/.config/systemd/user/kiosk-appliance.service
```

Der Service startet:

```text
dbus-run-session
-> cage
-> scripts/start-browser.sh
```

Die Unit wird fuer `default.target` aktiviert, damit sie nach dem tty1-Autologin im systemd User Manager startet.

## Debugging

```bash
systemctl status getty@tty1.service
cat /etc/systemd/system/getty@tty1.service.d/kiosk-autologin.conf
systemctl --user status kiosk-appliance.service
journalctl --user -u kiosk-appliance.service -f
pgrep -a cage
pgrep -a chromium
loginctl
```

## Desktop Edition

Die bestehende Desktop Edition bleibt erhalten. Sie verwendet die vorhandenen Module fuer Display Manager, native Sessions und Desktop-Fallbacks. Die Appliance Edition ist ein zusaetzliches Installationsprofil und entfernt keine bestehenden Pakete.
