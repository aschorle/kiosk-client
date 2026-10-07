# Scripts

Produktive Skripte:

- `browser-supervisor.sh`
- `start-browser.sh`
- `start-cage.sh`
- `start-wayland-session.sh`

`start-cage.sh` startet Cage mit `start-wayland-session.sh` als Client. Das Session-Skript versucht einen optional konfigurierten `DISPLAY_OUTPUT`/`DISPLAY_MODE`-Override erst nach Verfuegbarkeit des Wayland-Sockets und startet danach immer den `browser-supervisor.sh`. Ohne `DISPLAY_MODE` bleibt die automatische Moduswahl erhalten. Fehler beim Setzen des Modus werden protokolliert, blockieren den Kioskstart aber nicht. Der Supervisor startet `start-browser.sh`, ueberwacht Chromium und startet es bei Reload, Neustart oder Crash innerhalb der laufenden Cage-Sitzung neu.

## Linux source package

Linux-Installationsarchive werden niemals aus dem Windows-Arbeitsbaum erzeugt.
Der kanonische Packaging-Schritt archiviert einen expliziten Git-Commit mit
deaktivierter Autocrlf-Konvertierung und kann den Archivinhalt verifizieren:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/package-linux-source.ps1 `
  -Ref HEAD -OutputPath C:\Temp\source.zip -Verify
```

`scripts/test-linux-package.ps1` erzeugt ein temporäres Archiv und prüft, dass
alle gepackten `.sh`-Dateien LF verwenden, die Management-only-Installation
vorhanden ist und der Archivinhalt exakt dem Git-Commit entspricht.
