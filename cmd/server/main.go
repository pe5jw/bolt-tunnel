package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"tailscale.com/tsnet"
)

//go:embed gui.html
var guiHTML []byte

var (
	hostname   = flag.String("hostname", "bolttunnel-server", "naam op het tailnet")
	addr       = flag.String("addr", ":7780", "tailnet luisteradres")
	clientHost = flag.String("client", "bolttunnel-client", "hostname van de client")
	clientPort = flag.String("cport", "7781", "poort van de client /ping")
	guiAddr    = flag.String("gui", "localhost:8081", "lokaal adres voor de web-GUI")
	interval   = flag.Duration("interval", time.Second, "meet-interval")
	tunnels         = flag.String("tunnels", "", "server tunnel definitie: naam:poort of naam:tailnetpoort:host:poort")
	clientTunnels   = flag.String("client-tunnels", "", "tunnel configuratie voor de client: naam:clientpoort:serverpoort")
	commands        = flag.String("commands", "", "commando definitie voor client en server: naam:script")
	serverCommands  = flag.String("server-commands", "", "commando definitie alleen voor server GUI: naam:script")
)

// ── Monitor ──────────────────────────────────────────────────────────────────

type Sample struct {
	Time      int64   `json:"t"`
	LatencyMS float64 `json:"ms"`
	Direct    bool    `json:"dir"`
}

type Status struct {
	mu       sync.RWMutex
	samples  []Sample
	lost     int
	total    int
	dropouts int
	lastGood time.Time
}

func (s *Status) add(ms float64, direct bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.total++
	if ms < 0 {
		s.lost++
		if time.Since(s.lastGood) > 3*time.Second {
			s.dropouts++
		}
	} else {
		s.lastGood = time.Now()
	}
	samp := Sample{Time: time.Now().UnixMilli(), LatencyMS: ms, Direct: direct}
	if len(s.samples) >= 120 {
		s.samples = append(s.samples[1:], samp)
	} else {
		s.samples = append(s.samples, samp)
	}
}

type StatusSnapshot struct {
	Samples  []Sample `json:"samples"`
	AvgMS    float64  `json:"avg_ms"`
	JitterMS float64  `json:"jitter_ms"`
	LossPct  float64  `json:"loss_pct"`
	Dropouts int      `json:"dropouts"`
	Direct   bool     `json:"direct"`
	LED      string   `json:"led"`
}

func (s *Status) snapshot() StatusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := StatusSnapshot{
		Samples:  append([]Sample(nil), s.samples...),
		Dropouts: s.dropouts,
		LED:      "unknown",
	}
	if s.total > 0 {
		snap.LossPct = float64(s.lost) / float64(s.total) * 100
	}
	valid := make([]float64, 0, len(s.samples))
	for _, sm := range s.samples {
		if sm.LatencyMS >= 0 {
			valid = append(valid, sm.LatencyMS)
			if sm.Direct {
				snap.Direct = true
			}
		}
	}
	if len(valid) > 0 {
		sum := 0.0
		for _, v := range valid {
			sum += v
		}
		snap.AvgMS = sum / float64(len(valid))
		variance := 0.0
		for _, v := range valid {
			d := v - snap.AvgMS
			variance += d * d
		}
		snap.JitterMS = math.Sqrt(variance / float64(len(valid)))
		last := valid[len(valid)-1]
		switch {
		case last < 50:
			snap.LED = "good"
		case last < 150:
			snap.LED = "warn"
		default:
			snap.LED = "bad"
		}
	}
	return snap
}

var startTime = time.Now()

// ── Tunnel status ─────────────────────────────────────────────────────────────

type TunnelStatus struct {
	Name       string  `json:"name"`
	TailnetPort string `json:"tailnet_port"`
	Target     string  `json:"target"`
	Up         bool    `json:"up"`
	LatencyMS  float64 `json:"latency_ms"`
}

var (
	srvTunnelsMu  sync.RWMutex
	srvTunnels    []TunnelStatus
)

func getSrvTunnels() []TunnelStatus {
	srvTunnelsMu.RLock()
	defer srvTunnelsMu.RUnlock()
	out := make([]TunnelStatus, len(srvTunnels))
	copy(out, srvTunnels)
	return out
}

// ── Main ─────────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()

	srv := &tsnet.Server{
		Hostname:  *hostname,
		AuthKey:   os.Getenv("TS_AUTHKEY"),
		Logf:      func(string, ...any) {},
		Ephemeral: false,
	}
	defer srv.Close()

	if err := srv.Start(); err != nil {
		log.Fatalf("tsnet start mislukt: %v", err)
	}

	lc, err := srv.LocalClient()
	if err != nil {
		log.Fatalf("LocalClient mislukt: %v", err)
	}

	ctx := context.Background()
	for {
		st, err := lc.Status(ctx)
		if err == nil && st.Self != nil && len(st.Self.TailscaleIPs) > 0 {
			log.Printf("bolttunnel-server actief als %s (%s)", *hostname, st.Self.TailscaleIPs[0])
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// ── Monitor: latency naar de client ──────────────────────────────────────
	status := &Status{lastGood: time.Now()}
	tsHTTP := &http.Client{
		Transport: &http.Transport{DialContext: srv.Dial},
		Timeout:   4 * time.Second,
	}
	clientURL := "http://" + *clientHost + ":" + *clientPort

	go func() {
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()
		for range ticker.C {
			t0 := time.Now()
			resp, err := tsHTTP.Get(clientURL + "/ping")
			rtt := float64(time.Since(t0).Microseconds()) / 1000.0
			if err != nil {
				status.add(-1, false)
				continue
			}
			resp.Body.Close()
			st, _ := lc.Status(ctx)
			direct := false
			if st != nil {
				for _, p := range st.Peer {
					if p.HostName == *clientHost && p.CurAddr != "" {
						direct = true
					}
				}
			}
			status.add(rtt, direct)
		}
	}()

	// ── Tailnet listener: endpoints voor de client ────────────────────────────
	ln, err := srv.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("tailnet Listen mislukt: %v", err)
	}
	defer ln.Close()

	tailnetMux := http.NewServeMux()

	tailnetMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ts":   time.Now().UnixMilli(),
			"host": *hostname,
		})
	})

	tailnetMux.HandleFunc("/diag", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		st, err := lc.Status(ctx)
		diag := map[string]any{
			"os":     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"uptime": time.Since(startTime).Seconds(),
		}
		if err != nil {
			diag["ok"] = false
			diag["error"] = err.Error()
			json.NewEncoder(w).Encode(diag)
			return
		}
		diag["ok"] = true
		if st.Self != nil && len(st.Self.TailscaleIPs) > 0 {
			diag["tailscale_ip"] = st.Self.TailscaleIPs[0].String()
		}
		diag["backend"] = st.BackendState
		peers := []map[string]any{}
		for _, p := range st.Peer {
			peers = append(peers, map[string]any{
				"name":   p.HostName,
				"direct": p.CurAddr != "",
				"online": p.Online,
			})
		}
		diag["peers"] = peers
		snap := status.snapshot()
		diag["rtt_to_client_ms"] = snap.AvgMS
		checks := []map[string]any{
			{"name": "Node geregistreerd", "ok": diag["tailscale_ip"] != nil, "value": diag["tailscale_ip"]},
			{"name": "Backend", "ok": st.BackendState == "Running", "value": st.BackendState},
			{"name": "Peers", "ok": len(peers) > 0, "value": fmt.Sprintf("%d", len(peers))},
			{"name": "RTT naar client", "ok": snap.AvgMS > 0 && snap.AvgMS < 150, "value": fmt.Sprintf("%.1f ms", snap.AvgMS)},
		}
		diag["checks"] = checks
		json.NewEncoder(w).Encode(diag)
	})

	// ── Commando endpoints ───────────────────────────────────────────────────
	// Vaste lijst van toegestane commando's — geen vrije invoer mogelijk
	cmdMap := map[string]string{}
	if *commands != "" {
		for _, def := range strings.Split(*commands, ",") {
			parts := strings.SplitN(strings.TrimSpace(def), ":", 2)
			if len(parts) == 2 {
				cmdMap[parts[0]] = parts[1]
				log.Printf("Commando beschikbaar: %s → %s", parts[0], parts[1])
			}
		}
	}

	// /config — pusht client configuratie naar de client
	tailnetMux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Client tunnels
		clientTunnelList := []string{}
		if *clientTunnels != "" {
			for _, t := range strings.Split(*clientTunnels, ",") {
				if t = strings.TrimSpace(t); t != "" {
					clientTunnelList = append(clientTunnelList, t)
				}
			}
		}
		// Client commando's
		cmdNames := []string{}
		for name := range cmdMap {
			cmdNames = append(cmdNames, name)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"client_tunnels":  clientTunnelList,
			"client_commands": cmdNames,
		})
	})

	// /cmd/list — geeft lijst van beschikbare commando's terug
	tailnetMux.HandleFunc("/cmd/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		names := make([]string, 0, len(cmdMap))
		for name := range cmdMap {
			names = append(names, name)
		}
		json.NewEncoder(w).Encode(names)
	})

	// /cmd/run?name=xxx — voert een voorgedefinieerd commando uit
	tailnetMux.HandleFunc("/cmd/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		script, ok := cmdMap[name]
		if !ok {
			http.Error(w, `{"error":"onbekend commando"}`, 400)
			return
		}
		log.Printf("Commando uitvoeren: %s → %s", name, script)
		cmd := exec.Command("cmd", "/C", script)
		out, err := cmd.CombinedOutput()
		result := map[string]any{
			"name":   name,
			"output": string(out),
			"ok":     err == nil,
		}
		if err != nil {
			result["error"] = err.Error()
		}
		json.NewEncoder(w).Encode(result)
	})

	go func() {
		log.Printf("tailnet endpoints actief op %s", *addr)
		log.Fatal(http.Serve(ln, tailnetMux))
	}()

	// ── Tunnels op basis van -tunnels flag ───────────────────────────────────
	// Formaat A: naam:poort              → tailnet:poort → localhost:poort
	// Formaat B: naam:tailnetpoort:poort → tailnet:tailnetpoort → localhost:poort
	// Formaat C: naam:tailnetpoort:host:poort → tailnet:tailnetpoort → host:poort (LAN device)
	//
	// Voorbeelden:
	//   SDRoxide:4950                        → localhost:4950
	//   RDP:3389                             → localhost:3389
	//   Tuner-web:80:192.168.1.50:80         → 192.168.1.50:80
	//   Tuner-data:60001:192.168.1.50:60001  → 192.168.1.50:60001
	if *tunnels != "" {
		for _, def := range strings.Split(*tunnels, ",") {
			def = strings.TrimSpace(def)
			parts := strings.SplitN(def, ":", 4)

			var name, tailnetPort, targetAddr string

			switch len(parts) {
			case 2:
				// naam:poort → localhost:poort
				name = parts[0]
				tailnetPort = parts[1]
				targetAddr = "127.0.0.1:" + parts[1]
			case 3:
				// naam:tailnetpoort:lokaalpoort → localhost:lokaalpoort
				name = parts[0]
				tailnetPort = parts[1]
				targetAddr = "127.0.0.1:" + parts[2]
			case 4:
				// naam:tailnetpoort:host:poort → host:poort
				name = parts[0]
				tailnetPort = parts[1]
				targetAddr = parts[2] + ":" + parts[3]
			default:
				log.Printf("[WARN] Ongeldige tunnel definitie: %s", def)
				continue
			}

			tLn, err := srv.Listen("tcp", ":"+tailnetPort)
			if err != nil {
				log.Printf("[WARN] Tunnel %s: kan niet luisteren op tailnet poort %s: %v", name, tailnetPort, err)
				continue
			}
			log.Printf("Tunnel actief: %s tailnet:%s → %s", name, tailnetPort, targetAddr)

			// Registreer tunnel voor status tracking
			srvTunnelsMu.Lock()
			srvTunnels = append(srvTunnels, TunnelStatus{
				Name:        name,
				TailnetPort: tailnetPort,
				Target:      targetAddr,
			})
			idx := len(srvTunnels) - 1
			srvTunnelsMu.Unlock()

			// Probe goroutine: check elke 5 seconden of de target bereikbaar is
			go func(i int, target string) {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					t0 := time.Now()
					conn, err := net.DialTimeout("tcp", target, 3*time.Second)
					ms := float64(time.Since(t0).Microseconds()) / 1000.0
					srvTunnelsMu.Lock()
					if err == nil {
						conn.Close()
						srvTunnels[i].Up = true
						srvTunnels[i].LatencyMS = ms
					} else {
						srvTunnels[i].Up = false
						srvTunnels[i].LatencyMS = 0
					}
					srvTunnelsMu.Unlock()
				}
			}(idx, targetAddr)

			go func(name, targetAddr string) {
				for {
					client, err := tLn.Accept()
					if err != nil {
						return
					}
					go func(c net.Conn) {
						defer c.Close()
						target, err := net.Dial("tcp", targetAddr)
						if err != nil {
							log.Printf("Tunnel %s niet bereikbaar (%s): %v", name, targetAddr, err)
							return
						}
						defer target.Close()
						done := make(chan struct{}, 2)
						go func() { io.Copy(target, c); done <- struct{}{} }()
						go func() { io.Copy(c, target); done <- struct{}{} }()
						<-done
					}(client)
				}
			}(name, targetAddr)
		}
	}

	// ── Lokale GUI ────────────────────────────────────────────────────────────
	guiMux := http.NewServeMux()

	guiMux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		snap := status.snapshot()
		// Voeg tunnel status toe aan de snapshot
		type snapWithTunnels struct {
			Samples   interface{}    `json:"samples"`
			AvgMS     float64        `json:"avg_ms"`
			JitterMS  float64        `json:"jitter_ms"`
			LossPct   float64        `json:"loss_pct"`
			Dropouts  int            `json:"dropouts"`
			Direct    bool           `json:"direct"`
			LED       string         `json:"led"`
			Tunnels   []TunnelStatus `json:"tunnels"`
		}
		json.NewEncoder(w).Encode(snapWithTunnels{
			Samples:  snap.Samples,
			AvgMS:    snap.AvgMS,
			JitterMS: snap.JitterMS,
			LossPct:  snap.LossPct,
			Dropouts: snap.Dropouts,
			Direct:   snap.Direct,
			LED:      snap.LED,
			Tunnels:  getSrvTunnels(),
		})
	})

	// Server GUI: toon zowel gedeelde als server-only commando's
	srvCmdMap := map[string]string{}
	for k, v := range cmdMap { srvCmdMap[k] = v }
	if *serverCommands != "" {
		for _, def := range strings.Split(*serverCommands, ",") {
			parts := strings.SplitN(strings.TrimSpace(def), ":", 2)
			if len(parts) == 2 {
				srvCmdMap[parts[0]] = parts[1]
				log.Printf("Server-only commando: %s → %s", parts[0], parts[1])
			}
		}
	}

	guiMux.HandleFunc("/api/cmd/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		names := make([]string, 0, len(srvCmdMap))
		for name := range srvCmdMap {
			names = append(names, name)
		}
		json.NewEncoder(w).Encode(names)
	})

	guiMux.HandleFunc("/api/cmd/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		script, ok := srvCmdMap[name]
		if !ok {
			http.Error(w, `{"error":"onbekend commando"}`, 400)
			return
		}
		log.Printf("Commando uitvoeren: %s → %s", name, script)
		cmd := exec.Command("cmd", "/C", script)
		out, err := cmd.CombinedOutput()
		result := map[string]any{
			"name":   name,
			"output": string(out),
			"ok":     err == nil,
		}
		if err != nil {
			result["error"] = err.Error()
		}
		json.NewEncoder(w).Encode(result)
	})

	guiMux.HandleFunc("/api/diag", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp, err := tsHTTP.Get(clientURL + "/diag")
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
			return
		}
		defer resp.Body.Close()
		buf := make([]byte, 32*1024)
		w.WriteHeader(resp.StatusCode)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
					w.Write(buf[:n])
			}
			if readErr != nil {
				break
			}
		}
	})

	guiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(guiHTML)
	})

	log.Printf("Server GUI beschikbaar op http://%s", *guiAddr)
	log.Fatal(http.ListenAndServe(*guiAddr, guiMux))
}
