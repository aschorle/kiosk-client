Hardware-Notes
==============

Zielhardware

- Primär: Radxa Rock 4C+
- Sekundär: Raspberry Pi 4

Designprinzip

Hardware-spezifische Unterschiede werden minimal gehalten und auf Installer-Ebene gekapselt. Der Laufzeit-Client sollte ohne Board-spezifischen Code auskommen.

Wichtige Unterschiede und Hinweise

Radxa Rock 4C+
- AArch64-CPU, benötigt ggf. andere Firmware/Bootloader-Optionen
- Prüfen, ob Debian Bookworm Images oder Anpassungen nötig sind
- WLAN-Module und Treiber: je nach Revision ggf. proprietäre Firmware
- Thermisches Verhalten: Rock 4C+ kann je nach Last stärker wärmen — auf geringe CPU-Last achten

Raspberry Pi 4
- Weit verbreitete Unterstützung in Debian-basierten Images
- Proprietäre Firmware/Boot-Mechanismen beachten (boot partition)
- WLAN und GPU-Unterstützung in Standard-Images gut

Boot- und Kernel-Überlegungen

- Ziel ist ein Standard-Debian-Userland mit minimalen zusätzlichen Paketen
- Kernel- oder Boot-anpassungen sollten im Installer dokumentiert und optional gehalten werden

Empfehlung

- Testgeräte für beide Hardwaretypen einrichten
- Installer-Skripte so erstellen, dass sie die Hardware erkennen und die passenden Schritte ausführen
