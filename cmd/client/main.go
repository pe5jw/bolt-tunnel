package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"tailscale.com/tsnet"
)

//go:embed gui.html
var guiHTML []byte

var builtinAuthKey string

var (
	hostname     = flag.String("hostname", "bolttunnel-client", "naam op het tailnet")
	serverHost   = flag.String("server", "bolttunnel-server", "hostname van de server")
	serverPort   = flag.String("port", "7780", "poort van de server")
	listenAddr   = flag.String("listen", ":7781", "tailnet luisteradres voor server pings")
	guiAddr      = flag.String("gui", "127.0.0.1:8080", "lokaal adres voor de web-GUI")
	interval     = flag.Duration("interval", time.Second, "meet-interval")
	tunnels      = flag.String("tunnels", "", "extra tunnels: naam:clientpoort:serverpoort")
	useServerCfg = flag.Bool("server-config", true, "haal tunnel config op van de server")
	nowindow     = flag.Bool("nowindow", false, "geen GUI venster")
	logfile      = flag.String("logfile", "", "schrijf log naar bestand (leeg = alleen terminal)")
)

// ── Monitor ──────────────────────────────────────────────────────────────────

type Sample struct {
	Time      int64   `json:"t"`
	LatencyMS float64 `json:"ms"`
	Direct    bool    `json:"dir"`
}

type TunnelStatus struct {
	Name       string  `json:"name"`
	LocalPort  string  `json:"local_port"`
	ServerPort string  `json:"server_port"`
	Up         bool    `json:"up"`
	LatencyMS  float64 `json:"latency_ms"`
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
	Samples  []Sample       `json:"samples"`
	AvgMS    float64        `json:"avg_ms"`
	JitterMS float64        `json:"jitter_ms"`
	LossPct  float64        `json:"loss_pct"`
	Dropouts int            `json:"dropouts"`
	Direct   bool           `json:"direct"`
	LED      string         `json:"led"`
	Tunnels  []TunnelStatus `json:"tunnels"`
}

var (
	tunnelsMu      sync.RWMutex
	tunnelStatuses []TunnelStatus
)

func logStartup(msg string) {
	log.Println(msg)
}

func (s *Status) snapshot() StatusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tunnelsMu.RLock()
	ts := make([]TunnelStatus, len(tunnelStatuses))
	copy(ts, tunnelStatuses)
	tunnelsMu.RUnlock()

	snap := StatusSnapshot{
		Samples:  append([]Sample(nil), s.samples...),
		Dropouts: s.dropouts,
		LED:      "unknown",
		Tunnels:  ts,
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

func probeTunnels(ctx context.Context, srv *tsnet.Server, serverHost string, defs []string) {
	tunnelsMu.Lock()
	for _, def := range defs {
		parts := strings.Split(strings.TrimSpace(def), ":")
		if len(parts) != 3 {
			continue
		}
		tunnelStatuses = append(tunnelStatuses, TunnelStatus{
			Name:       parts[0],
			LocalPort:  parts[1],
			ServerPort: parts[2],
		})
	}
	tunnelsMu.Unlock()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tunnelsMu.Lock()
			for i := range tunnelStatuses {
				t0 := time.Now()
				conn, err := srv.Dial(ctx, "tcp", serverHost+":"+tunnelStatuses[i].ServerPort)
				ms := float64(time.Since(t0).Microseconds()) / 1000.0
				if err == nil {
					conn.Close()
					tunnelStatuses[i].Up = true
					tunnelStatuses[i].LatencyMS = ms
				} else {
					tunnelStatuses[i].Up = false
					tunnelStatuses[i].LatencyMS = 0
				}
			}
			tunnelsMu.Unlock()
		}
	}
}

// startTunnel start een TCP listener en tunnelt naar de server
func startTunnel(ctx context.Context, srv *tsnet.Server, def string, remoteHost string) {
	parts := strings.Split(strings.TrimSpace(def), ":")
	if len(parts) != 3 {
		log.Printf("[WARN] Ongeldige tunnel definitie: %s", def)
		return
	}
	name, localPort, remotePort := parts[0], parts[1], parts[2]
	ln, err := net.Listen("tcp", "127.0.0.1:"+localPort)
	if err != nil {
		log.Printf("[WARN] Tunnel %s: poort %s bezet: %v", name, localPort, err)
		return
	}
	logStartup(fmt.Sprintf("✓ Tunnel actief: %s localhost:%s → server:%s", name, localPort, remotePort))
	for {
		client, err := ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			server, err := srv.Dial(ctx, "tcp", remoteHost+":"+remotePort)
			if err != nil {
				return
			}
			defer server.Close()
			done := make(chan struct{}, 2)
			go func() { io.Copy(server, c); done <- struct{}{} }()
			go func() { io.Copy(c, server); done <- struct{}{} }()
			<-done
		}(client)
	}
}

// ── Main ─────────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()

	// Log naar bestand als opgegeven
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("[WARN] Kan logbestand niet openen: %v", err)
		} else {
			log.SetOutput(f)
			log.Printf("Bolttunnel client gestart - log naar %s", *logfile)
		}
	}

	authKey := builtinAuthKey
	if authKey == "" {
		authKey = os.Getenv("TS_AUTHKEY")
	}

	srv := &tsnet.Server{
		Hostname: *hostname,
		AuthKey:  authKey,
		Logf:     func(string, ...any) {},
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

	// ── Auth flow ─────────────────────────────────────────────────────────────
	authDone := make(chan struct{})
	go func() {
		for {
			st, err := lc.Status(ctx)
			if err == nil && st.Self != nil && len(st.Self.TailscaleIPs) > 0 {
				logStartup(fmt.Sprintf("✓ Verbonden met tailnet als %s (%s)", *hostname, st.Self.TailscaleIPs[0]))
				close(authDone)
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	select {
	case <-authDone:
	case <-time.After(3 * time.Second):
		authURL := ""
		for authURL == "" {
			st, _ := lc.Status(ctx)
			if st != nil && st.AuthURL != "" {
				authURL = st.AuthURL
			}
			time.Sleep(300 * time.Millisecond)
		}
		logStartup("⚠ Tailscale login vereist - open in browser:")
		logStartup("▷ " + authURL)
		log.Printf("Tailscale auth vereist - open in browser: %s", authURL)
		exec.Command("cmd", "/C", "start", authURL).Start()
		<-authDone
	}

	// ── HTTP client via tailnet ───────────────────────────────────────────────
	tsHTTP := &http.Client{
		Transport: &http.Transport{DialContext: srv.Dial},
		Timeout:   4 * time.Second,
	}
	serverURL := "http://" + *serverHost + ":" + *serverPort

	// ── Extra tunnels direct starten ──────────────────────────────────────────
	if *tunnels != "" {
		for _, t := range strings.Split(*tunnels, ",") {
			if t = strings.TrimSpace(t); t != "" {
				go startTunnel(ctx, srv, t, *serverHost)
			}
		}
	}

	// ── Server config asynchroon ophalen ──────────────────────────────────────
	if *useServerCfg {
		go func() {
			logStartup("▷ Server config ophalen van " + serverURL)
			for i := 0; ; i++ {
				resp, err := tsHTTP.Get(serverURL + "/config")
				if err == nil {
					var cfg struct {
						ClientTunnels []string `json:"client_tunnels"`
					}
					if json.NewDecoder(resp.Body).Decode(&cfg) == nil {
						resp.Body.Close()
						logStartup(fmt.Sprintf("✓ Server config ontvangen: %d tunnels", len(cfg.ClientTunnels)))
						for _, t := range cfg.ClientTunnels {
							if t = strings.TrimSpace(t); t != "" {
								go startTunnel(ctx, srv, t, *serverHost)
							}
						}
						go probeTunnels(ctx, srv, *serverHost, cfg.ClientTunnels)
						return
					}
					resp.Body.Close()
				}
				if i < 30 {
					time.Sleep(2 * time.Second)
				} else {
					time.Sleep(30 * time.Second)
				}
			}
		}()
	}

	// ── Monitor ───────────────────────────────────────────────────────────────
	status := &Status{lastGood: time.Now()}

	go func() {
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()
		for range ticker.C {
			t0 := time.Now()
			resp, err := tsHTTP.Get(serverURL + "/ping")
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
					if p.HostName == *serverHost && p.CurAddr != "" {
						direct = true
					}
				}
			}
			status.add(rtt, direct)
		}
	}()

	// ── Tailnet listener ──────────────────────────────────────────────────────
	ln, err := srv.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("tailnet Listen mislukt: %v", err)
	}
	defer ln.Close()

	pingMux := http.NewServeMux()
	pingMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ts":` + fmt.Sprintf("%d", time.Now().UnixMilli()) + `,"host":"` + *hostname + `"}`))
	})
	go func() { http.Serve(ln, pingMux) }()

	// ── Lokale GUI ────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		snap := status.snapshot()
json.NewEncoder(w).Encode(snap)
	})

	mux.HandleFunc("/api/diag", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp, err := tsHTTP.Get(serverURL + "/diag")
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	})

	mux.HandleFunc("/api/cmd/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp, err := tsHTTP.Get(serverURL + "/cmd/list")
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	})

	mux.HandleFunc("/api/cmd/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Query().Get("name")
		resp, err := tsHTTP.Get(serverURL + "/cmd/run?name=" + name)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(guiHTML)
	})

	guiLn, err := net.Listen("tcp", *guiAddr)
	if err != nil {
		log.Fatalf("Kan niet luisteren op %s: %v", *guiAddr, err)
	}
	go http.Serve(guiLn, mux)
	log.Printf("GUI beschikbaar op http://%s", *guiAddr)

	// ── Browser openen ───────────────────────────────────────────────────────
	if !*nowindow {
		url := "http://" + *guiAddr
		log.Printf("GUI beschikbaar op %s", url)
		exec.Command("cmd", "/C", "start", url).Start()
	}
	// Blijf draaien
	select {}
}
