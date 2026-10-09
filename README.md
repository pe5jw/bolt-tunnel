# tailmon — zero-client tailnet verbindingsmonitor

Geen Tailscale installatie nodig. Beide kanten (client en server) embedden
`tsnet` en registreren zichzelf als node op het tailnet.

## Architectuur

```
Client machine                    Tailscale control          Server / subnet router
┌─────────────────────┐          ┌────────────────┐         ┌──────────────────────────┐
│  tailmon-client      │◄──auth──►│ login.tailscale│◄──auth──│  tailmon-server           │
│  ┌───────────────┐  │          └────────────────┘         │  ┌────────────────────┐   │
│  │ tsnet.Server  │  │                                      │  │ tsnet.Server       │   │
│  │ (embedded TS) │◄─┼─── WireGuard P2P (of DERP) ────────►│  │ (embedded TS)      │   │
│  └───────────────┘  │                                      │  └────────────────────┘   │
│  ┌───────────────┐  │          Elke seconde:               │  ┌────────────────────┐   │
│  │ monitor loop  │──┼── GET /ping ──────────────────────── ►│  │ GET /ping → {ts}  │   │
│  │ meet RTT      │◄─┼── 200 OK ──────────────────────────  │  └────────────────────┘   │
│  └───────────────┘  │                                      │  ┌────────────────────┐   │
│  ┌───────────────┐  │          Op verzoek:                 │  │ GET /diag → json  │   │
│  │ web GUI       │  │── GET /diag ──────────────────────── ►│  └────────────────────┘   │
│  │ localhost:8080│  │                                      │  ┌────────────────────┐   │
│  └───────────────┘  │                                      │  │ → Remote LAN       │   │
└─────────────────────┘                                      │  │   192.168.x.x/24   │   │
                                                             └──────────────────────────┘
```

## Voordelen van tsnet

- **Geen Tailscale installatie** nodig op client of server
- **Geen root/admin rechten** vereist
- **Geen systeem-daemon** (`tailscaled`) nodig
- **Eén binary** per kant — makkelijk te distribueren
- De verbinding **gaat altijd via het tailnet** — ook de `/ping` meting

## Setup

### 1. Tailscale auth key aanmaken

Ga naar https://login.tailscale.com/admin/settings/keys en maak een
reusable auth key aan (of ephemeral voor éénmalig gebruik).

### 2. Server deployen (op de subnet router / gateway)

```bash
cd cmd/server
go build -o tailmon-server .

# Met auth key als env variabele (aanbevolen):
TS_AUTHKEY=tskey-auth-xxxx ./tailmon-server \
  --hostname tailmon-server \
  --addr :7780

# Of zonder key — eerste run opent een auth URL in de terminal:
./tailmon-server
```

### 3. Client draaien

```bash
cd cmd/client
go build -o tailmon-client .

TS_AUTHKEY=tskey-auth-xxxx ./tailmon-client \
  --hostname tailmon-client \
  --server tailmon-server \
  --gui localhost:8080

# Open de GUI:
open http://localhost:8080
```

## Endpoints (server)

| Endpoint | Methode | Doel | Gewicht |
|----------|---------|------|---------|
| `/ping`  | GET     | Latency meting — minimale response | ~100 bytes |
| `/diag`  | GET     | Uitgebreide diagnostiek — peers, DERP, checks | ~2 KB |

## Endpoints (client, lokaal)

| Endpoint     | Doel |
|--------------|------|
| `/api/status` | Live status snapshot — de GUI pollt dit elke seconde |
| `/api/diag`  | Proxy naar server `/diag` — alleen op verzoek |
| `/`          | Web GUI |

## De GUI embedden in je eigen app

Als je de monitor wilt embedden in een bestaande Go app:

```go
import "github.com/yourname/tailmon/monitor"

// Start de monitor (één keer, bij app start)
mon := monitor.New(monitor.Config{
    TSNet:      yourTsnetServer,  // je bestaande tsnet.Server
    ServerHost: "tailmon-server",
    Interval:   time.Second,
})
mon.Start(ctx)

// Status LED ophalen (non-blocking, geen latency impact)
snap := mon.Snapshot()
// snap.LED → "good" | "warn" | "bad" | "unknown"
// snap.AvgMS → gemiddelde latency
// snap.LossPct → packet loss %

// HTTP handlers registreren
mux.Handle("/api/monitor/status", mon.StatusHandler())
mux.Handle("/api/monitor/diag",   mon.DiagHandler())
```

## Latency impact van de monitor zelf

De monitor doet één HTTP GET per seconde via het al bestaande
WireGuard-pad. Dit is verwaarloosbaar:
- Geen extra verbinding — hergebruikt het bestaande tsnet pad
- Request body: 0 bytes
- Response body: ~60 bytes JSON
- CPU: < 0.1% op een moderne CPU

## Integratie in bestaande Go app (zonder aparte client binary)

Als je app al tsnet gebruikt, voeg je alleen de monitor goroutine toe:

```go
// Eén goroutine, één HTTP client, één ticker
go func() {
    client := &http.Client{
        Transport: &http.Transport{DialContext: tsnetSrv.Dial},
        Timeout: 4 * time.Second,
    }
    for range time.NewTicker(time.Second).C {
        t0 := time.Now()
        resp, err := client.Get("http://tailmon-server:7780/ping")
        rtt := time.Since(t0)
        if err != nil {
            status.Add(-1)
        } else {
            resp.Body.Close()
            status.Add(rtt.Seconds() * 1000)
        }
    }
}()
```
