#!/bin/bash
# Bolttunnel server starter
# Pas de variabelen hieronder aan

HOSTNAME="bolttunnel-server"
PORT=7780
CLIENT="bolttunnel-client"
CLIENT_PORT=7781
GUI="localhost:8081"

# Server tunnels (naam:poort of naam:tailnetpoort:host:poort voor LAN apparaten)
TUNNELS="SDRoxide:4950,BoltSDR:6443,BoltDVK:4532,TCI40001:40001,RDP:3389,SWITCH:9090,BoltATR1000:3000"

# Client tunnels — worden naar alle clients gepusht (naam:clientpoort:serverpoort)
CLIENT_TUNNELS="SDRoxide:4951:4950,BoltSDR:6444:6443,BoltDVK:4533:4532,TCI40001:40001:40001,SWITCH:9091:9090,BoltATR1000:9092:3000"

# ─────────────────────────────────────────────────────────────────────────────

# Detecteer architectuur
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  BIN="./bolttunnel-server-linux-amd64" ;;
  aarch64) BIN="./bolttunnel-server-linux-arm64" ;;
  armv7l)  BIN="./bolttunnel-server-linux-armv7" ;;
  *)       echo "[FOUT] Onbekende architectuur: $ARCH"; exit 1 ;;
esac

if [ ! -f "$BIN" ]; then
    echo "[FOUT] Binary niet gevonden: $BIN"
    echo "       Download de juiste versie van https://github.com/pe5jw/bolt-tunnel/releases"
    exit 1
fi

chmod +x "$BIN"

echo "Bolttunnel server starten als \"$HOSTNAME\""
echo "  Tailnet poort  : $PORT"
echo "  Client meten   : $CLIENT:$CLIENT_PORT"
echo "  GUI            : http://$GUI"
echo "  Tunnels        : $TUNNELS"
echo ""

"$BIN" \
  --hostname "$HOSTNAME" \
  --addr ":$PORT" \
  --client "$CLIENT" \
  --cport "$CLIENT_PORT" \
  --gui "$GUI" \
  --tunnels "$TUNNELS" \
  --client-tunnels "$CLIENT_TUNNELS"
