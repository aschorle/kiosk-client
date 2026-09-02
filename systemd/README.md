# systemd

Produktive User-Units:

- `kiosk-agent.service`
- `kiosk-appliance.service`

Beide Units laufen im systemd User Manager des Kiosk-Benutzers und werden fuer `default.target` aktiviert.

Fuer Management-only-Clients gibt es getrennt die Systemunit
`kiosk-agent-management-only.service` und die eng begrenzte sudoers-Vorlage
`sudoers.d/kiosk-agent-mini-browser`. Sie sind ausschliesslich fuer den
bestehenden Mini-PC-Browserdienst `kiosk.service` vorgesehen und werden nicht
vom Appliance-Installer verwendet.
