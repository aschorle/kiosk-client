# Appliance Edition

Die Appliance Edition installiert den kiosk-client ohne Display Manager und ohne Desktop Environment. GNOME, KDE, GDM, SDDM oder LightDM werden nicht installiert, konfiguriert oder entfernt.

## Zielarchitektur

```text
Boot
-> systemd getty@tty1
-> Autologin des Kiosk-Benutzers
-> systemd --user default.target
-> kiosk-agent.service
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
sudo KIOSK_USER=rock ./installer/install.sh
```

Wenn `KIOSK_USER` nicht gesetzt ist, verwendet der Installer `SUDO_USER`.
`installer/install.sh` erkennt das Board und delegiert an den passenden
Appliance-Installer. Direkte Aufrufe von `installer/appliance.sh` bleiben fuer
Tests des generischen Appliance-Profils moeglich.

## Installierte Pakete

Die Appliance Edition installiert nur:

- `ca-certificates`
- `chromium`
- `cage`
- `dbus`
- `dbus-user-session`

Es werden keine Display-Manager- oder Desktop-Pakete installiert.

## tty Autologin

`installer/tty.sh` schreibt eine systemd-Override-Datei fuer `getty@tty1`:

```text
/etc/systemd/system/getty@tty1.service.d/kiosk-autologin.conf
```

Diese meldet den Kiosk-Benutzer automatisch auf tty1 an. GDM, SDDM und LightDM bleiben unberuehrt.

## Runtime

`installer/runtime.sh` installiert die User Services:

```text
~/.config/systemd/user/kiosk-agent.service
~/.config/systemd/user/kiosk-appliance.service
```

`kiosk-agent.service` startet die lokale Administrationsoberflaeche und API.
`kiosk-appliance.service` startet:

```text
dbus-run-session
-> cage
-> scripts/start-browser.sh
```

Beide Units werden fuer `default.target` aktiviert, damit sie nach dem
tty1-Autologin im systemd User Manager starten.

## Debugging

```bash
systemctl status getty@tty1.service
cat /etc/systemd/system/getty@tty1.service.d/kiosk-autologin.conf
systemctl --user status kiosk-agent.service
systemctl --user status kiosk-appliance.service
journalctl --user -u kiosk-agent.service -f
journalctl --user -u kiosk-appliance.service -f
pgrep -a cage
pgrep -a chromium
loginctl
```

## Desktop Edition

Die bestehende Desktop Edition bleibt eingefroren. Sie verwendet die vorhandenen
Legacy-Module fuer Display Manager, native Sessions und Desktop-Fallbacks. Die
Appliance Edition fuehrt diese Module nicht aus und entfernt keine bestehenden
Pakete.
