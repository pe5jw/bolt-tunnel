#!/bin/bash
# Bolttunnel client starter
# Pas de variabelen hieronder aan

HOSTNAME="bolttunnel-client"
SERVER="bolttunnel-server"
PORT=7780
LISTEN=7781
GUI="127.0.0.1:8080"

# Extra tunnels alleen voor deze client (naam:clientpoort:serverpoort)
# Laat leeg als je geen extra tunnels nodig hebt
EXTRA_TUNNELS="RDP:13389:3389"

# Logbestand (laat leeg voor alleen scherm)
LOGFILE=""

# ─────────────────────────────────────────────────────────────────────────────

# Detecteer architectuur
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  BIN="./bolttunnel-client-linux-amd64" ;;
  aarch64) BIN="./bolttunnel-client-linux-arm64" ;;
  armv7l)  BIN="./bolttunnel-client-linux-armv7" ;;
  *)       echo "[FOUT] Onbekende architectuur: $ARCH"; exit 1 ;;
esac

if [ ! -f "$BIN" ]; then
    echo "[FOUT] Binary niet gevonden: $BIN"
    echo "       Download de juiste versie van https://github.com/pe5jw/bolt-tunnel/releases"
    exit 1
fi

chmod +x "$BIN"

ARGS="--hostname $HOSTNAME --server $SERVER --port $PORT --listen :$LISTEN --gui $GUI --nowindow"

if [ -n "$EXTRA_TUNNELS" ]; then
    ARGS="$ARGS --tunnels \"$EXTRA_TUNNELS\""
fi

if [ -n "$LOGFILE" ]; then
    ARGS="$ARGS --logfile $LOGFILE"
fi

echo "Bolttunnel client starten..."
echo "  Hostname : $HOSTNAME"
echo "  Server   : $SERVER:$PORT"
echo "  GUI      : http://$GUI"
echo ""

eval "$BIN $ARGS"
