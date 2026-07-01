# Appliance Edition

Die Appliance Edition ist der einzige unterstuetzte Produktpfad.

## Zielarchitektur

```text
Boot
-> systemd getty@tty1
-> Autologin des Kiosk-Benutzers
-> systemd --user default.target
-> kiosk-agent.service
-> kiosk-appliance.service
-> dbus-run-session
-> scripts/start-cage.sh
-> cage
-> scripts/browser-supervisor.sh
-> scripts/start-browser.sh
-> Chromium
-> URL aus config/client.conf
```

## Installation

```bash
sudo KIOSK_USER=rock ./installer/install.sh
```

## Runtime-Dateien

```text
~/.config/systemd/user/kiosk-agent.service
~/.config/systemd/user/kiosk-appliance.service
/etc/systemd/system/getty@tty1.service.d/kiosk-autologin.conf
```

## Pakete

- `ca-certificates`
- `chromium`
- `cage`
- `dbus`
- `dbus-user-session`
- `fonts-noto-color-emoji`

## Betrieb

```bash
systemctl --user status kiosk-agent.service
systemctl --user status kiosk-appliance.service
journalctl --user -u kiosk-agent.service -f
journalctl --user -u kiosk-appliance.service -f
```
