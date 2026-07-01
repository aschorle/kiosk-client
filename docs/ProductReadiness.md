# Product Readiness

Stand: Version 0.12.3

Der produktive Pfad ist ausschliesslich die Appliance Edition fuer Debian oder Armbian Minimal.

## Aktuelle Architektur

```text
Boot
-> systemd
-> getty Autologin
-> systemd --user default.target
-> kiosk-agent.service
-> kiosk-appliance.service
-> dbus-run-session
-> cage
-> scripts/start-browser.sh
-> Chromium
-> konfigurierte URL
```

## Audit-Ergebnis

### ENTFERNT

- alte grafische Login-Module aus `installer/`
- alte Session- und Service-Module aus `installer/`
- alte Runtime-Unit `systemd/user/kiosk-runtime.service`
- alte direkte Browser-Unit
- Doku zu historischen grafischen Installationspfaden
- Paket- und Doku-Verweise auf alte Browser-Service-Steuerung

### ERSETZT

- Browsersteuerung steuert jetzt `kiosk-appliance.service`
- Reload startet die Appliance-Runtime neu
- Config-Speichern ist vom Browser-Neustart getrennt
- Installation beschreibt nur noch den Appliance-Pfad
- systemd-Dokumentation beschreibt nur noch User-Units fuer `default.target`

### BEIBEHALTEN

- `chromium-browser` als Erkennungsname: wird fuer Debian-/Armbian-Images mit abweichendem Binary-Namen benoetigt.
- `scripts/start-cage.sh`: wird von der Appliance-Runtime direkt verwendet.
- Board-Installer fuer Radxa und Raspberry Pi: delegieren auf das gemeinsame Appliance-Profil.
- `installer/install-common.sh`: gemeinsame Pruef- und Logging-Funktionen.
- `installer/verify.sh`: read-only Vorabpruefungen.
- `installer/packages.sh`: zentrale Appliance-Paketliste.

## Offene Punkte bis 1.0

- Installation auf frischen Bookworm- und Trixie-Minimalsystemen erneut verifizieren.
- Release-Artefakt mit vorgebautem `kiosk-agent` definieren.
- Admin-Oberflaeche final polieren.
- Dokumentation fuer Betrieb und Update abschliessen.
