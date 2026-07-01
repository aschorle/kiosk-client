# CHANGELOG

## 0.12.7

- Browser-Reload und Browser-Neustart beenden jetzt Chromium innerhalb der laufenden Cage-Sitzung statt `kiosk-appliance.service` direkt neu zu starten.

## 0.12.6

- Board-Erkennung fuer Raspberry Pi 3 Model B Rev 1.2 ergaenzt; der bestehende Raspberry-Pi-Installer bleibt zustaendig.

## 0.12.5

- Transparentes Xcursor-Theme korrigiert: gueltige 24x24-Cursor-Datei mit volltransparenten ARGB-Pixeln und zusaetzlichen Cursor-Aliasnamen.

## 0.12.4

- Mauszeiger in der Appliance-Runtime ueber ein lokales transparentes Xcursor-Theme ausgeblendet.

## 0.12.3

- Appliance-Installer korrigiert Ownership fuer erzeugte Benutzer-Konfigurationspfade unter `~/.config`.

## 0.12.2

- Chromium-Startparameter fuer Chromium 149 auf Debian Trixie angepasst: Crashpad/Breakpad-Deaktivierung entfernt.

## 0.12.1

- Appliance-Installer installiert Go automatisch, wenn `kiosk-agent` fehlt und kein Go vorhanden ist.
- Vorhandene ausfuehrbare `kiosk-agent`-Binaries werden weiterhin unveraendert verwendet.

## 0.12.0

- Chromium-Flag `--disable-gpu` entfernt, damit GPU-Beschleunigung auf RK3399/Panfrost getestet werden kann.
- Release-Audit vor Version 1.0: veraltete Dokumentation bereinigt.
- Projektstruktur auf Appliance Edition fuer Debian/Armbian Minimal reduziert.
