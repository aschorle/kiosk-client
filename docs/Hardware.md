# Hardware

Primaere Zielplattform:

- Radxa ROCK 4C+ / RK3399

Weitere getestete Zielplattform:

- Raspberry Pi 4

## Produktiver Referenzclient

- Hostname: `raspi-kiosk`
- Hardware: Raspberry Pi 4 Model B
- Betriebssystem: Armbian 26.5.1 / Debian trixie
- Kernel: `6.18.35-current-bcm2711`
- Architektur: AArch64
- SSH-Alias: `raspi-kiosk` (Benutzer `aschorle`, dedizierter Key
  `~/.ssh/id_ed25519_raspi_kiosk`)

## Laufzeitannahmen

- Debian oder Armbian Minimal
- AArch64
- systemd
- Cage
- Chromium

Board-spezifische Unterschiede bleiben im Installer gekapselt. Die Agent- und Web-Logik enthaelt keine Board-spezialisierten Pfade.
