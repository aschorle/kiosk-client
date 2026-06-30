Browser
=======

Zweck

Der kiosk-client startet Chromium als einzige sichtbare Anwendung. In der aktuellen Phase wird ausschließlich das Browser-Startskript umgesetzt. Wayland, Cage, systemd, Webinterface und die endgültige URL-Konfiguration folgen später.

Startskript

Das Skript `scripts/start-browser.sh` sucht Chromium automatisch über `command -v`. Unterstützt werden die Programmnamen `chromium` und `chromium-browser`.

Die Standard-URL ist:

```text
http://localhost
```

Das Skript ist bereits darauf vorbereitet, später `config/client.conf` zu lesen. Wenn dort eine einfache Zeile `URL=...` vorhanden ist, wird diese URL verwendet. Wenn die Datei fehlt oder keine URL enthält, bleibt `http://localhost` der Fallback.

Chromium-Parameter

- `--kiosk`: startet Chromium im Vollbild-Kioskmodus ohne normale Browser-Oberfläche.
- `--incognito`: startet ohne persistente Sitzungshistorie und reduziert lokale Spuren zwischen Läufen.
- `--no-first-run`: unterdrückt Einrichtungsdialoge beim ersten Start.
- `--disable-session-crashed-bubble`: verhindert Wiederherstellungsdialoge nach einem vorherigen unsauberen Abbruch.
- `--disable-infobars`: reduziert Browser-Hinweisleisten, die den Kiosk-Inhalt überdecken könnten.
- `--disable-features=Translate`: deaktiviert automatische Übersetzungsfunktionen und zugehörige Einblendungen.
- `--disable-sync`: deaktiviert Chromium-Synchronisierung, da der Kiosk-Client keine Benutzerprofile synchronisieren soll.
- `--overscroll-history-navigation=0`: deaktiviert Navigation durch Overscroll-Gesten, damit die angezeigte Webanwendung nicht versehentlich verlassen wird.

Bewusste Abgrenzung

Das Startskript setzt noch keine Wayland- oder Cage-Parameter, richtet keinen Autostart ein und erzeugt keinen systemd-Service. Es konfiguriert außerdem keine Browser-Policies und keinen Cache. Diese Themen werden getrennt umgesetzt, damit Browserstart, Display-Stack und Service-Betrieb einzeln testbar bleiben.
