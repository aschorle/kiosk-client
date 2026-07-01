# Scripts

Produktive Skripte:

- `browser-supervisor.sh`
- `start-browser.sh`
- `start-cage.sh`

`start-cage.sh` startet Cage. Cage startet `browser-supervisor.sh`. Der Supervisor startet `start-browser.sh`, ueberwacht Chromium und startet es bei Reload, Neustart oder Crash innerhalb der laufenden Cage-Sitzung neu.
