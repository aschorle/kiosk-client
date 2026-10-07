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
-> scripts/start-wayland-session.sh (optional: configured output mode via wlr-randr)
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
- `wlr-randr`
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

## Optionaler Wayland-Ausgabemodus

`DISPLAY_OUTPUT` benennt den Wayland-Ausgang (zum Beispiel `HDMI-A-1`), auf
dem `DISPLAY_MODE` gesetzt werden soll. `DISPLAY_MODE` enthält den von
`wlr-randr` akzeptierten Modus, zum Beispiel `1680x1050@59.883Hz`. Beide
Parameter sind optional. Ohne `DISPLAY_MODE` führt der Start kein
`wlr-randr` aus und Cage/DRM wählt den Modus wie bisher automatisch; ein
gesetzter Modus benötigt zusätzlich einen passenden `DISPLAY_OUTPUT`.

Beispiel für den ROCK 4C+ mit dem getesteten ANMITE-Display in
`config/client.conf`:

```ini
DISPLAY_OUTPUT=HDMI-A-1
DISPLAY_MODE=1680x1050@59.883Hz
```

Der Agent bewahrt diese lokalen Werte bei Konfigurationsänderungen über das
Dashboard auf. Die Werte werden nicht als Dashboard-/API-Konfiguration
offengelegt.

Der Installer installiert `wlr-randr` als Runtime-Abhängigkeit. Der Modus wird
erst innerhalb der gestarteten Cage-Sitzung angefordert, nachdem deren
Wayland-Socket verfügbar ist. Fehlt `wlr-randr`, der konfigurierte Ausgang oder
Modus wird nicht akzeptiert, oder ist der Socket nicht rechtzeitig verfügbar,
erscheint eine Warnung. Browser-Supervisor und Chromium starten trotzdem;
Cage/DRM behalten ihre automatische Moduswahl bei. Der Installer erzwingt
keinen Modus anhand des Boards und ändert keine Display-Konfiguration
automatisch.

Raspberry-Installationen ohne Override bleiben unverändert bei automatischer
Moduswahl. Beim getesteten Raspberry/ANMITE-Aufbau wird weiterhin nativ
1920x1200 gewählt.

Den aktiven Ausgang und Modus in der laufenden Benutzersitzung prüfen:

```bash
wlr-randr
```

Der aktive Modus ist mit `(current)` markiert. Zum Beispiel wurde auf dem
ROCK 4C+ nach dem kontrollierten Service-Neustart und nach einem vollständigen
Reboot jeweils `1680x1050 px, 59.882999 Hz (current)` bestätigt. Der
Hardware-Fallbacktest mit einem ungültigen Modus erzeugte die erwartete
Warnung; Cage, Browser-Supervisor und Chromium liefen weiter. Nach dem Reboot
waren `kiosk-agent.service` und `kiosk-appliance.service` aktiv und der
konfigurierte Modus automatisch wieder gesetzt.

### Fehlersuche bei nicht unterstützten Modi

- `wlr-randr` in der Cage-Sitzung ausführen und den exakten Ausgangsnamen sowie
  die angebotenen Modi prüfen. Namen und Schreibweise in `client.conf` müssen
  dazu passen.
- Sicherstellen, dass `DISPLAY_MODE` und `DISPLAY_OUTPUT` gemeinsam gesetzt
  sind und dass der Modus tatsächlich für diesen Ausgang angeboten wird.
- `journalctl --user -u kiosk-appliance.service` auf Warnungen zum
  Wayland-Socket, zu `wlr-randr` oder zur Modusauswahl prüfen.
- Nicht unterstützte Modi werden nicht erzwungen. Den Override auskommentieren
  oder `DISPLAY_MODE` leeren, um zur automatischen Auswahl zurückzukehren;
  ein Fehlschlag des Overrides soll den Kioskstart nicht verhindern.
