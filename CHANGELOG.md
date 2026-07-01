# CHANGELOG

## 0.12.2

- Chromium-Startparameter fuer Chromium 149 auf Debian Trixie angepasst: Crashpad/Breakpad-Deaktivierung entfernt.

## 0.12.1

- Appliance-Installer installiert Go automatisch, wenn `kiosk-agent` fehlt und kein Go vorhanden ist.
- Vorhandene ausfuehrbare `kiosk-agent`-Binaries werden weiterhin unveraendert verwendet.

## 0.12.0

- Chromium-Flag `--disable-gpu` entfernt, damit GPU-Beschleunigung auf RK3399/Panfrost getestet werden kann.
- Release-Audit vor Version 1.0: veraltete Dokumentation bereinigt.
- Projektstruktur auf Appliance Edition fuer Debian/Armbian Minimal reduziert.
