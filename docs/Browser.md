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

Das Skript lädt `config/client.conf`. Wenn die Datei fehlt oder nicht lesbar ist, bricht der Browserstart mit einer verständlichen Fehlermeldung ab. Wenn `URL` leer ist, bleibt `http://localhost` der Fallback.

Konfigurationsdatei

Die Datei `config/client.conf` enthält die Laufzeitkonfiguration des Clients:

```text
URL=http://localhost
DEVICE_ID=display01
BROWSER=chromium
```

- `URL`: Adresse der Webseite, die im Kioskmodus geöffnet wird. Wenn der Wert leer ist, verwendet das Startskript `http://localhost`.
- `DEVICE_ID`: Eindeutige Kennung des Geräts. Sie wird vom Browser-Startskript noch nicht ausgewertet, ist aber Teil der zentralen Client-Konfiguration.
- `BROWSER`: Name des Chromium-Programms, das über `command -v` gesucht wird. Standard ist `chromium`; als Fallback wird `chromium-browser` akzeptiert.

Chromium-Parameter

- `--kiosk`: startet Chromium im Vollbild-Kioskmodus ohne normale Browser-Oberfläche.
- `--incognito`: startet ohne persistente Sitzungshistorie und reduziert lokale Spuren zwischen Läufen.
- `--no-first-run`: unterdrückt Einrichtungsdialoge beim ersten Start.
- `--disable-session-crashed-bubble`: verhindert Wiederherstellungsdialoge nach einem vorherigen unsauberen Abbruch.
- `--disable-infobars`: reduziert Browser-Hinweisleisten, die den Kiosk-Inhalt überdecken könnten.
- `--disable-gpu`: deaktiviert die GPU-Beschleunigung. Das reduziert auf der Radxa Rock 4C+ die Abhängigkeit von Treiber- und Compositor-Verhalten.
- `--disable-crash-reporter`: deaktiviert den Chromium-Crash-Reporter, damit im Kiosk-Betrieb keine zusätzlichen Reporter-Prozesse oder Dialogpfade entstehen.
- `--disable-breakpad`: deaktiviert Breakpad, die von Chromium genutzte Crash-Erfassung.
- `--disable-background-networking`: verhindert automatische Hintergrund-Netzwerkaktivität des Browsers, die für den reinen Kiosk-Client nicht benötigt wird.
- `--disable-background-timer-throttling`: verhindert, dass Chromium Timer von Hintergrundseiten drosselt. Das hält Webanwendungen stabiler, wenn Chromium Fensterzustände intern als Hintergrundzustand bewertet.
- `--disable-renderer-backgrounding`: verhindert, dass Renderer-Prozesse bei Hintergrundbewertung herunterpriorisiert werden.
- `--disable-sync`: deaktiviert Chromium-Synchronisierung, da der Kiosk-Client keine Benutzerprofile synchronisieren soll.
- `--overscroll-history-navigation=0`: deaktiviert Navigation durch Overscroll-Gesten, damit die angezeigte Webanwendung nicht versehentlich verlassen wird.

Stabilisierung in Version 0.3.3

Die Runtime verwendet bewusst keine Wayland- oder Ozone-Parameter. Auf der Radxa Rock 4C+ bleibt Chromium in dieser Version im X11-Pfad, weil die installierte Chromium-Version dort stabiler betrieben werden kann.

Experimentelle Feature-Schalter werden vermieden. Die Parameter beschränken sich auf Kioskmodus, Dialogunterdrückung, deaktivierte Hintergrundfunktionen und eine konservative Grafikstrategie ohne GPU-Beschleunigung.

Bewusste Abgrenzung

Das Startskript setzt noch keine Wayland- oder Cage-Parameter, richtet keinen Autostart ein und erzeugt keinen systemd-Service. Es konfiguriert außerdem keine Browser-Policies und keinen Cache. Diese Themen werden getrennt umgesetzt, damit Browserstart, Display-Stack und Service-Betrieb einzeln testbar bleiben.
