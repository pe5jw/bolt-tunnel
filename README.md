# Bolttunnel

Bolttunnel maakt beveiligde TCP-tunnels via een Tailscale tailnet, zonder dat Tailscale geïnstalleerd hoeft te worden op de client of server. Gebouwd voor Windows en Linux/Raspberry Pi, ideaal voor SDR-ontvangers, TCI-streams en andere lokale services.

## Kenmerken

- **Geen Tailscale installatie nodig** — werkt via embedded [tsnet](https://pkg.go.dev/tailscale.com/tsnet)
- **TCP-tunnels** naar lokale services en LAN-apparaten (configureerbaar via flags of server-config)
- **Verbindingsmonitor GUI** in de browser — latency, jitter, packet loss, dropouts, sparkline
- **Drie tabs**: Status | Verbinding (diagnostics) | Beheer (server commando's)
- **Minimaliseerknop** in de statusbalk — inklapbaar met ▲/▼
- **Server beheer** — start/stop/herstart scripts op de server via de client GUI
- Werkt op **Windows, Linux en Raspberry Pi** (arm64)

## Snel starten

### Download
Ga naar [Releases](https://github.com/pe5jw/bolt-tunnel/releases) en download de juiste versie:

| Platform | Bestand |
|---|---|
| Windows | bolttunnel-windows.zip |
| Linux x86_64 | bolttunnel-linux-amd64.zip |
| Raspberry Pi 4/5 | bolttunnel-linux-arm64.zip |

### Vereisten
- Een [Tailscale](https://tailscale.com) account (gratis tier volstaat)
- Een auth key van https://login.tailscale.com/admin/settings/keys

### Installatie en gebruik
Zie [HANDLEIDING.md](HANDLEIDING.md) voor volledige instructies, inclusief:
- Tailscale account en auth key aanmaken
- Server en client instellen
- Meerdere servers
- Linux/Pi installatie met systemd

## Bouwen vanuit broncode

Vereisten: [Go 1.21+](https://go.dev/dl/) (of laat `setup.bat` dit automatisch installeren)

```bat
git clone https://github.com/pe5jw/bolt-tunnel.git
cd bolt-tunnel
build.bat
```

Linux:
```bash
git clone https://github.com/pe5jw/bolt-tunnel.git
cd bolt-tunnel
cd cmd/server && go build -o ../../bin/bolttunnel-server .
cd ../client  && go build -o ../../bin/bolttunnel-client .
```

## Architectuur

```
[Client PC]                          [Server / Pi]
bolttunnel-client.exe                bolttunnel-server
  tsnet (tailscale embedded)   ←→    tsnet (tailscale embedded)
  localhost:POORT ←→ TCP-tunnel ←→   localhost:DOEL_POORT
  GUI: http://127.0.0.1:8080         GUI: http://localhost:8081
```

Beide processen doen mee aan het tailnet zonder Tailscale daemon. Tunnels lopen volledig versleuteld over WireGuard.

## Licentie

MIT — zie [LICENSE](LICENSE).  
Bevat [tailscale.com/tsnet](https://pkg.go.dev/tailscale.com/tsnet) onder BSD 3-Clause licentie.

---

[English README](README-EN.md)
