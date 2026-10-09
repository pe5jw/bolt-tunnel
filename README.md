# Bolttunnel

Bolttunnel maakt een beveiligde tunnel via een Tailscale tailnet, zonder dat Tailscale geïnstalleerd hoeft te zijn op de client of server. Ideaal voor het op afstand bereiken van services zoals SDR-ontvangers, TCI-streams en andere lokale apparaten.

## Kenmerken

- Geen Tailscale installatie nodig — werkt via embedded tsnet
- TCP tunnels naar lokale services en LAN-apparaten
- Verbindingsmonitor GUI in de browser (latency, jitter, packet loss)
- Server beheer knoppen vanuit de client GUI
- Werkt op Windows, Linux en Raspberry Pi

## Snel starten

### Download
Ga naar [Releases](https://github.com/pe5jw/bolt-tunnel/releases) en download de juiste versie:

| Platform | Bestand |
|---|---|
| Windows | bolttunnel-windows.zip |
| Linux x86_64 | bolttunnel-linux-amd64.zip |
| Raspberry Pi 4/5 | bolttunnel-linux-arm64.zip |

### Vereisten
- Een [Tailscale](https://tailscale.com) account
- Een auth key van https://login.tailscale.com/admin/settings/keys

### Installatie en gebruik
Zie [HANDLEIDING.md](HANDLEIDING.md) voor volledige instructies, inclusief:
- Tailscale account en auth key aanmaken
- Server en client instellen
- Meerdere servers
- Linux/Pi installatie met systemd

## Bouwen vanuit broncode

Vereisten: [Go 1.21+](https://go.dev/dl/)

```bat
git clone https://github.com/pe5jw/bolt-tunnel.git
cd bolt-tunnel
build.bat
```


wget https://github.com/pe5jw/bolt-tunnel/releases/download/v1.0.0/install.sh
bash install.sh
cd ~/bolttunnel
nano start-client.sh   # hostname en server aanpassen
./start-client.sh
## Licentie

MIT
