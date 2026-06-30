# Product Readiness Analyse

Stand: Version 0.10.0. Diese Analyse bereitet Version 1.0 vor und beschreibt
den aktuellen technischen Zustand ohne funktionale Aenderungen.

## 1. Aktuelle Architektur

Der kiosk-client ist als lokale Appliance fuer genau ein Geraet ausgelegt:

- ein lokaler Agent mit REST-API und lokaler Weboberflaeche
- eine Konfigurationsdatei unter `config/client.conf`
- Chromium als einziger Browser
- Cage als minimale Wayland-Kiosk-Shell fuer die Appliance Edition
- systemd User Services fuer Agent, Browser- und Runtime-Prozesse
- Shell-Installer unter `installer/`

Aktueller Zielpfad fuer die Appliance Edition:

```text
Boot
-> systemd
-> getty Autologin
-> systemd --user
-> kiosk-appliance.service
-> dbus-run-session
-> cage
-> Chromium
-> konfigurierte URL
```

Die Desktop Edition existiert weiterhin als historischer Pfad, ist aber fuer
neue Produktarbeit eingefroren. Fuer Version 1.0 sollte der Installer nur noch
den Appliance-Pfad produktiv verwenden.

## 2. Installer-Analyse

| Datei | Aufgabe | Weiter benoetigt? | Einordnung | Bewertung |
| --- | --- | --- | --- | --- |
| `installer/install.sh` | Dispatcher mit Root-, Architektur-, Debian- und Board-Pruefung; delegiert an Board-Installer. | Ja, aber vereinfachen. | Allgemein | Sollte fuer 1.0 auf Appliance-Installation statt Desktop-/Board-Mischkette fuehren. |
| `installer/install-common.sh` | Logging, Debian-/Architektur-/Netzwerk-/Speicher-/Board-Pruefungen. | Ja. | Allgemein | Sinnvoller gemeinsamer Kern. Debian-Erkennung muss Armbian sauber beruecksichtigen. |
| `installer/packages.sh` | Zentrale Paketlisten. | Ja, aber bereinigen. | Allgemein | Enthalt aktuell Entwickler- und Desktop-nahe Komfortpakete in `COMMON_PACKAGES`. Fuer Appliance 1.0 zu gross. |
| `installer/appliance.sh` | Minimaler Appliance-Installer mit `chromium`, `cage`, `dbus`, danach `runtime.sh` und `tty.sh`. | Ja. | Appliance | Naechster Zielpfad fuer 1.0. Muss ggf. Agent-Service und Build/Deployment des Agent klaeren. |
| `installer/runtime.sh` | Installiert und aktiviert `kiosk-appliance.service` im systemd User-Kontext. | Ja. | Appliance | Zentral fuer Minimalpfad. Aktuell nur Appliance-Runtime, nicht Agent. |
| `installer/tty.sh` | Schreibt getty@tty1 Autologin Override fuer den Kiosk-Benutzer. | Ja. | Appliance | Passt zum Zielpfad ohne Display Manager. |
| `installer/browser.sh` | Installiert und verifiziert Chromium. | Optional/zusammenfuehren. | Appliance | Funktional sinnvoll, aber in `appliance.sh` teilweise doppelt als Paketliste enthalten. |
| `installer/cage.sh` | Installiert und verifiziert Cage. | Optional/zusammenfuehren. | Appliance | Funktional sinnvoll, aber in `appliance.sh` teilweise doppelt als Paketliste enthalten. |
| `installer/verify.sh` | Read-only Preflight Checks. | Ja. | Allgemein | Sinnvoll, sollte von Appliance-Installer konsistent verwendet werden. |
| `installer/install-radxa.sh` | Radxa-Installationskette mit Basisinstallation und Desktop-Modulen. | Ja, aber stark bereinigen. | Gemischt | Aktuell nicht 1.0-minimal, weil Desktop-Module ausgefuehrt werden. |
| `installer/install-rpi.sh` | Platzhalter fuer Raspberry Pi 4. | Optional. | Allgemein | Enthalt keine produktive Logik. Fuer 1.0 entweder implementieren oder aus produktivem Dispatcher entfernen. |
| `installer/autologin.sh` | GDM3 Autologin. | Nein fuer Appliance. | Desktop | Desktop-spezifisch, fuer 1.0 nicht im Appliance-Pfad verwenden. |
| `installer/session.sh` | Display-Manager-Session fuer GDM/SDDM/LightDM. | Nein fuer Appliance. | Desktop | Desktop-spezifisch und komplex. Nicht fuer Minimal-Installer verwenden. |
| `installer/power.sh` | GNOME gsettings plus systemd logind/sleep Drop-ins. | Teilweise. | Gemischt | GNOME-Teil entfernen/isolieren. systemd Sleep/Logind-Hardening kann appliance-relevant sein. |
| `installer/network.sh` | Platzhalter fuer NetworkManager/WLAN. | Nein im aktuellen Zustand. | Allgemein/Desktop-nahe | Keine produktive Logik. NetworkManager ist fuer kabelgebundene Minimal-Appliance nicht zwingend. |
| `installer/cleanup.sh` | Platzhalter fuer Cleanup. | Nein im aktuellen Zustand. | Allgemein | Keine produktive Logik. Erst behalten, wenn konkrete Cleanup-Regeln existieren. |
| `installer/wayland.sh` | Platzhalter fuer Wayland/Cage-Konfiguration. | Nein im aktuellen Zustand. | Allgemein | Keine produktive Logik. Cage wird bereits ueber Runtime gestartet. |

Kernaussage: Es gibt bereits einen guten Appliance-Strang (`appliance.sh`,
`runtime.sh`, `tty.sh`, `kiosk-appliance.service`). Der historische
Radxa-Strang ruft dagegen weiterhin Desktop-Module auf und ist fuer Version 1.0
als produktiver Minimal-Installer nicht geeignet.

## 3. Paketanalyse

| Paket | Quelle | Bewertung | Einordnung | Empfehlung |
| --- | --- | --- | --- | --- |
| `chromium` | `APPLIANCE_PACKAGES`, `KIOSK_PACKAGES`, `browser.sh` | Erforderlich | Appliance | Behalten. Kern des Produkts. |
| `cage` | `APPLIANCE_PACKAGES`, `KIOSK_PACKAGES`, `cage.sh` | Erforderlich | Appliance | Behalten. Minimale Kiosk-Shell. |
| `dbus` | `APPLIANCE_PACKAGES` | Erforderlich | Appliance | Behalten, weil `kiosk-appliance.service` `dbus-run-session` nutzt. |
| `ca-certificates` | `COMMON_PACKAGES` | Erforderlich | Allgemein | Behalten, damit HTTPS-Ziele und Paketquellen funktionieren. |
| `openssh-server` | `COMMON_PACKAGES` | Optional | Betrieb/Diagnose | Fuer Headless-Erstzugriff sinnvoll, aber als Produktentscheidung dokumentieren. |
| `network-manager` | `COMMON_PACKAGES` | Optional | Allgemein/Desktop-nahe | Entfernen, wenn Minimal-System mit systemd-networkd oder statischer Netzkonfiguration auskommt. Nur installieren, wenn WLAN/NetworkManager explizit Ziel ist. |
| `golang-go` | `COMMON_PACKAGES` | Optional/entfernen | Build | Fuer reproduzierbare Appliance besser nicht auf Zielsystem bauen. Stattdessen Binary ausliefern oder separaten Build-Schritt definieren. |
| `git` | `COMMON_PACKAGES` | Optional/entfernen | Installation/Entwicklung | Fuer Produktinstaller vermeiden, wenn Artefakt anderweitig bereitgestellt wird. |
| `curl` | `COMMON_PACKAGES` | Optional | Diagnose/Download | Nur behalten, wenn Installer es wirklich nutzt. Aktuell nicht zwingend. |
| `wget` | `COMMON_PACKAGES` | Optional/entfernen | Diagnose/Download | Redundant zu `curl`; nicht beide installieren. |
| `vim` | `COMMON_PACKAGES` | Entfernen | Komfort | Kein Appliance-Erfordernis. |
| `htop` | `DEV_PACKAGES` | Entfernen | Entwicklung/Diagnose | Nicht fuer Minimal-Installer. |
| `tree` | `DEV_PACKAGES` | Entfernen | Entwicklung/Diagnose | Nicht fuer Minimal-Installer. |

Offen ist eine Zielsystem-Pruefung der transitiven Abhaengigkeiten von
`chromium`, `cage` und `dbus` auf Debian und Armbian. Diese Analyse bewertet nur
die explizit vom Projekt angeforderten Pakete.

## 4. systemd-Analyse

| Unit | Aufgabe | Weiter benoetigt? | Einordnung | Bewertung |
| --- | --- | --- | --- | --- |
| `systemd/user/kiosk-appliance.service` | Startet `dbus-run-session -- cage -- start-browser.sh`. | Ja. | Appliance | Zentrale Runtime fuer 1.0. Passt zum Zielpfad. |
| `systemd/user/kiosk-agent.service` | Startet den lokalen Agent, aktuell an `graphical-session.target` gebunden. | Ja, aber anpassen. | Gemischt | Fuer Appliance noetig, aber `graphical-session.target` passt nicht zum getty/systemd-user-Pfad. |
| `systemd/user/kiosk-runtime.service` | Startet `scripts/start-cage.sh`, gebunden an `graphical-session.target`. | Nein fuer Appliance 1.0. | Desktop/Legacy | Als Legacy/Fallback dokumentieren, aber nicht im Minimalpfad installieren/aktivieren. |
| `systemd/user/kiosk-browser.service` | Startet Chromium direkt ohne Cage, gebunden an `graphical-session.target`. | Nein fuer Appliance 1.0. | Desktop/Legacy | Nur Fallback, nicht produktiv installieren/aktivieren. |

Offener Punkt: Der Appliance-Installer installiert derzeit `kiosk-appliance.service`,
aber nicht den Agent-Service. Fuer Version 1.0 muss entschieden werden, ob der
Agent als eigener User Service unter `default.target` laeuft oder ob eine andere
saubere systemd-Struktur verwendet wird. Dabei duerfen Browser- und Agent-Logik
nicht vermischt werden.

## 5. Desktop-Abhaengigkeiten

Desktop-spezifisch und fuer die Appliance Edition nicht produktiv relevant:

- `installer/autologin.sh`: GDM3 Autologin
- `installer/session.sh`: GDM, SDDM, LightDM, AccountsService, Wayland-Session-Datei fuer Display Manager
- GNOME-Teil in `installer/power.sh`: `gsettings`, GNOME Screensaver, GNOME Power Settings
- `systemd/user/kiosk-runtime.service`: `graphical-session.target`
- `systemd/user/kiosk-browser.service`: direkter Browser-Fallback in grafischer Desktop-Session
- `systemd/user/kiosk-agent.service` in aktueller Form: Bindung an `graphical-session.target`
- Dokumentierte alte Pfade rund um GDM/GNOME/KDE/Plasma/X11/Display Manager

Diese Bestandteile sollten nicht mehr durch den produktiven Installer fuer 1.0
ausgefuehrt werden. Wenn sie im Repository bleiben, sollten sie klar als Legacy
oder Desktop Edition markiert sein.

## 6. Appliance-Abhaengigkeiten

Fuer den Minimalpfad weiterhin erforderlich:

- systemd System Manager
- systemd User Manager
- getty@tty1 Autologin
- existierender Kiosk-Benutzer
- `dbus-run-session`
- Cage
- Chromium
- `scripts/start-browser.sh`
- `config/client.conf`
- lokaler Agent und Weboberflaeche, wenn Administration nach Installation verfuegbar sein soll

Appliance-relevante Dateien:

- `installer/appliance.sh`
- `installer/runtime.sh`
- `installer/tty.sh`
- `installer/install-common.sh`
- `installer/verify.sh`
- `systemd/user/kiosk-appliance.service`
- `scripts/start-browser.sh`
- optional: `scripts/start-cage.sh`, falls der Service nicht direkt Cage startet

## 7. Empfohlene Bereinigungen

1. Produktiven Installer-Einstieg festlegen: `installer/appliance.sh` als
   Zielpfad oder `installer/install.sh` so umbauen, dass er auf den Appliance-
   Pfad delegiert.
2. `installer/install-radxa.sh` von Desktop-Modulen entkoppeln oder fuer 1.0
   aus dem produktiven Pfad entfernen.
3. `COMMON_PACKAGES` auf wirklich notwendige Pakete reduzieren.
4. Zielsystem-Build vermeiden: `golang-go` aus dem Appliance-Paketset entfernen,
   sofern ein reproduzierbares Binary ausgeliefert wird.
5. Komfort- und Entwicklungswerkzeuge aus dem Produktinstaller entfernen:
   `vim`, `htop`, `tree`, ggf. `git`, `wget`.
6. NetworkManager nur installieren, wenn die Appliance-Netzwerkstrategie ihn
   ausdruecklich benoetigt.
7. Agent-Service fuer Appliance sauber definieren, ohne
   `graphical-session.target`.
8. Desktop-Module nicht mehr automatisch ausfuehren:
   `autologin.sh`, `session.sh`, Desktop-Anteile aus `power.sh`.
9. Platzhaltermodule entweder mit konkretem Zweck fuellen oder aus der
   produktiven Installationskette entfernen.
10. Installer-Dokumentation aktualisieren, damit sie nicht mehr alte
    Desktop-Zwischenstaende als produktiven Weg beschreibt.

## 8. Offene Punkte bis Version 1.0

- Debian- und Armbian-Erkennung finalisieren. Armbian kann `ID=debian`,
  `ID=armbian` oder gemischte `/etc/os-release` Werte liefern.
- Unterstuetzte Zielversionen festlegen, mindestens Debian 12/Bookworm und die
  passende Armbian-Basis.
- Reproduzierbaren Agent-Build definieren: auf Zielsystem bauen oder Binary
  ausliefern. Fuer Minimalismus ist Binary-Auslieferung vorzuziehen.
- Kiosk-Benutzer-Strategie klaeren: vorhandener Benutzer, `KIOSK_USER`, oder
  vom Installer angelegter dedizierter Benutzer.
- Agent-Service fuer Appliance unter `default.target` oder gleichwertigem
  User-Target entwerfen und installieren.
- Entscheiden, ob `scripts/start-cage.sh` im Appliance-Pfad weiterhin noetig
  ist, da `kiosk-appliance.service` Cage bereits direkt startet.
- Paketliste auf frischem Debian-Minimal und Armbian-Minimal testen.
- Sicherstellen, dass keine Desktop-Pakete transitiv oder direkt durch eigene
  Paketlisten angefordert werden.
- Installationslauf idempotent testen: erster Lauf, zweiter Lauf, Reboot.
- Fehlerfaelle dokumentieren: kein Netzwerk, fehlender Benutzer, fehlendes
  `config/client.conf`, fehlendes Chromium/Cage.

## 9. Empfohlene Reihenfolge der letzten Entwicklungsschritte

1. Appliance-Installer als einzigen produktiven Zielpfad festlegen.
2. Paketliste minimalisieren und dokumentieren.
3. Agent-Installation und Agent-systemd-Unit fuer Appliance abschliessen.
4. Kiosk-Benutzer und getty Autologin reproduzierbar machen.
5. Debian-/Armbian-Erkennung finalisieren.
6. Radxa- und Raspberry-Pi-Unterschiede nur dort abbilden, wo sie wirklich
   noetig sind.
7. Desktop-/Legacy-Module aus der produktiven Kette entfernen.
8. Frische Minimalinstallation auf Debian testen.
9. Frische Minimalinstallation auf Armbian testen.
10. Dokumentation und Version-1.0-Checkliste abschliessen.

## Zusammenfassung

Der Codebestand enthaelt bereits die Bausteine fuer eine minimale Appliance.
Der wichtigste Produkt-Reife-Schritt ist nicht das Hinzufuegen neuer Funktionen,
sondern das Entfernen des alten Desktop-Installationspfads aus der produktiven
Kette und das Schliessen der Luecke zwischen Appliance-Runtime und lokalem
Agent-Service.
