# Bolttunnel

Bolttunnel creates secure TCP tunnels over a Tailscale tailnet without requiring Tailscale to be installed on either the client or server. Built for Windows and Linux/Raspberry Pi, ideal for SDR receivers, TCI streams, and other local services.

## Features

- **No Tailscale installation required** — uses embedded [tsnet](https://pkg.go.dev/tailscale.com/tsnet)
- **TCP tunnels** to local services and LAN devices (configurable via flags or server-pushed config)
- **Browser-based monitoring GUI** — latency, jitter, packet loss, dropouts, sparkline chart
- **Three tabs**: Status | Connection (diagnostics) | Management (server commands)
- **Minimize button** in the status bar — collapsible with ▲/▼
- **Server management** — run start/stop/restart scripts on the server from the client GUI
- Runs on **Windows, Linux and Raspberry Pi** (arm64)

## Quick start

### Download
Go to [Releases](https://github.com/pe5jw/bolt-tunnel/releases) and download the right version:

| Platform | File |
|---|---|
| Windows | bolttunnel-windows.zip |
| Linux x86_64 | bolttunnel-linux-amd64.zip |
| Raspberry Pi 4/5 | bolttunnel-linux-arm64.zip |

### Requirements
- A [Tailscale](https://tailscale.com) account (free tier is enough)
- An auth key from https://login.tailscale.com/admin/settings/keys

### Installation and usage
See [HANDLEIDING.md](HANDLEIDING.md) (Dutch) for full instructions, including:
- Creating a Tailscale account and auth key
- Configuring server and client
- Multiple servers
- Linux/Pi installation with systemd

## Build from source

Requirements: [Go 1.21+](https://go.dev/dl/) (or let `setup.bat` install it automatically)

Windows:
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

## Architecture

```
[Client PC]                          [Server / Pi]
bolttunnel-client.exe                bolttunnel-server
  tsnet (embedded tailscale)   ←→    tsnet (embedded tailscale)
  localhost:PORT ←→ TCP tunnel ←→    localhost:TARGET_PORT
  GUI: http://127.0.0.1:8080         GUI: http://localhost:8081
```

Both processes join the tailnet without a Tailscale daemon. All tunnel traffic is WireGuard-encrypted end-to-end.

## Command-line flags

### Client
| Flag | Default | Description |
|---|---|---|
| `--hostname` | `bolttunnel-client` | Node name on the tailnet |
| `--server` | `bolttunnel-server` | Server hostname on the tailnet |
| `--port` | `7780` | Server HTTP port |
| `--listen` | `:7781` | Tailnet listen address (for server pings) |
| `--gui` | `127.0.0.1:8080` | Local GUI address |
| `--interval` | `1s` | Latency measurement interval |
| `--tunnels` | | Extra tunnels: `name:localport:serverport,...` |
| `--server-config` | `true` | Fetch tunnel config from server |
| `--nowindow` | `false` | Don't open browser window |
| `--logfile` | | Write log to file |

### Server
| Flag | Default | Description |
|---|---|---|
| `--hostname` | `bolttunnel-server` | Node name on the tailnet |
| `--addr` | `:7780` | Tailnet listen address |
| `--client` | `bolttunnel-client` | Client hostname on the tailnet |
| `--gui` | `localhost:8081` | Local GUI address |
| `--tunnels` | | Tunnels: `name:tailnetport:localport` or `name:tailnetport:host:port` |
| `--client-tunnels` | | Tunnels to push to clients: `name:clientport:serverport,...` |
| `--commands` | | Allowed commands: `name:script,...` |

## Security notes

- Auth keys are passed at build time via `-ldflags "-X main.builtinAuthKey=..."` or at runtime via `TS_AUTHKEY` environment variable
- The build script (`build.bat`) prompts for the key interactively and clears it immediately after use — it is never written to disk
- Tailscale state (node keys, IPs) is stored locally in a `tsnet-state/` directory — this directory is excluded from git

## Changelog

### v1.1.0 (2026-10-11)
- GUI: three tabs — Status, Connection (diagnostics), Management (server commands)
- GUI: minimize button ▲/▼ in status bar, state persisted in localStorage
- Server management moved to its own tab (was in Status tab)
- Fixed go.mod module name
- Updated READMEs (NL + EN)

### v1.0.0
- Initial release

## License

MIT — see [LICENSE](LICENSE).  
Contains [tailscale.com/tsnet](https://pkg.go.dev/tailscale.com/tsnet) under BSD 3-Clause license.

---

[Nederlandse README](README.md)
