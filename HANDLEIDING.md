# Bolttunnel — Installatie & Tailnet handleiding

## 1. Tailscale account aanmaken
1. Ga naar https://tailscale.com en klik **Get started**
2. Maak een account aan (Google, GitHub of e-mail)
3. Je hebt nu een eigen tailnet

## 2. Auth key aanmaken
1. Ga naar https://login.tailscale.com/admin/settings/keys
2. Klik **Generate auth key**
3. Instellingen: Reusable=aan, Ephemeral=aan
4. Kopieer de key (begint met tskey-auth-...)
   WAARSCHUWING: Je ziet de key maar één keer!

## 3. Bouwen
Vereisten: Go 1.21+ van https://go.dev/dl/

Windows:
  build.bat  (vraagt om auth key, bouwt distro\ map)

Linux binary (vanuit PowerShell):
  $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/bolttunnel-client-linux ./cmd/client
  $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/bolttunnel-server-linux ./cmd/server
  $env:GOOS=""; $env:GOARCH=""

## 4. Server instellen
Pas start-server.bat aan:
  set HOSTNAME=bolttunnel-server
  set TUNNELS=SDRoxide:4950,BoltSDR:6443,RDP:3389
  set CLIENT_TUNNELS=SDRoxide:4951:4950,BoltSDR:6444:6443

## 5. Client instellen
Pas start-client.bat aan:
  set HOSTNAME=bolttunnel-client
  set SERVER=bolttunnel-server
  set EXTRA_TUNNELS=RDP:13389:3389

## 6. Meerdere servers
Kopie van start-client.bat met:
  set HOSTNAME=bolttunnel-client2
  set SERVER=bolttunnel-server2
  set LISTEN=7782
  set GUI=127.0.0.1:8082

## 7. Linux (zonder GUI)
  ./bolttunnel-client-linux --hostname bolttunnel-client --server bolttunnel-server --port 7780 --listen :7781 --gui 127.0.0.1:8080 --nowindow
  Open browser op http://127.0.0.1:8080

## Veiligheid
- Sla de auth key NOOIT op in een bat-bestand dat je deelt
- Trek gecompromitteerde keys in via https://login.tailscale.com/admin/settings/keys
