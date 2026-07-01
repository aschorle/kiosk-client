# CHANGELOG

## 0.13.6

- Repository-Audit vor Code Freeze durchgefuehrt.
- Veraltete Dokumentation zu Browsersteuerung, Chromium-Flags, Installer-Buildpfad und Appliance-Paketen bereinigt.

## 0.13.5

- `fonts-noto-color-emoji` in die Appliance-Paketliste aufgenommen.
- Abschlussmeldung des Installers finalisiert: kein automatischer Reboot, nur Hinweis auf `sudo reboot`.
- Sichtbarer Mauszeiger unter Wayland/Cage als bekannte plattformabhaengige Einschraenkung dokumentiert.

## 0.13.4

- Installer-Abschluss stabilisiert: nach erfolgreicher Appliance-Installation kein automatischer Reboot, sondern klarer Hinweis auf `sudo reboot`.
- Cursor-Ausblendung erneut geprueft: keine zusaetzlichen Workarounds; die Runtime bleibt beim vorhandenen transparenten Xcursor-Theme.

## 0.13.3

- Reboot-Berechtigung korrigiert: Installer richtet eine enge sudoers-Regel fuer `/usr/bin/systemctl reboot` ein, der Agent nutzt `sudo /usr/bin/systemctl reboot`.

## 0.13.2

- Appliance-Installer baut `kiosk-agent` bei jeder erfolgreichen Installation aus den aktuellen Repository-Quellen neu.

## 0.13.1

- Kiosk-Konfigurationsseite um `System Reboot` erweitert; neuer Agent-Endpunkt `POST /api/system/reboot` loest einen sauberen System-Reboot aus.

## 0.13.0

- Status-API und Kiosk-Konfigurationsseite zeigen die CPU-Temperatur aus `/sys/class/thermal/thermal_zone0/temp` an; fehlt der Wert, wird `n/a` ausgegeben.

## 0.12.9

- Browser-Supervisor fuer die Appliance-Runtime eingefuehrt: Reload und Neustart laufen per `SIGUSR1`/`SIGUSR2`, ohne Cage zu beenden.

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
