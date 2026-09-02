# Installer

Der Installer kennt nur den Appliance-Pfad.

Einstieg:

```bash
sudo KIOSK_USER=rock ./installer/install.sh
```

Aktive Module:

- `install-common.sh`
- `install.sh`
- `install-radxa.sh`
- `install-rpi.sh`
- `appliance.sh`
- `packages.sh`
- `verify.sh`
- `runtime.sh`
- `tty.sh`

## Management-only

`install-management-only.sh` ist ein separater, noch nicht ausgefuehrter Pfad
fuer Ubuntu 24.04 amd64. Er erwartet ein vorgebautes Agent-Artefakt mit
SHA-256-Pruefsumme und eine vorbereitete, tokenhaltige Konfiguration ausserhalb
des Repositories. Standardmaessig installiert er Dateien nur und aktiviert
keinen Service. Er darf weder LightDM, Xorg, LXDE, Openbox, Chromium,
`kiosk.service`, nginx, Filebrowser noch `/srv/kiosk` veraendern.
