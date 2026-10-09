#!/bin/bash
# Bolttunnel installatie script
# Downloadt de juiste binary voor deze machine en maakt scripts uitvoerbaar

set -e

REPO="pe5jw/bolt-tunnel"
VERSION="v1.0.0"
INSTALL_DIR="${1:-$HOME/bolttunnel}"

# Detecteer architectuur
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ZIPFILE="bolttunnel-linux-amd64.zip" ;;
  aarch64) ZIPFILE="bolttunnel-linux-arm64.zip" ;;
  armv7l)  ZIPFILE="bolttunnel-linux-arm64.zip"
           echo "[WARN] armv7 — arm64 zip wordt gebruikt, controleer binary naam" ;;
  *)       echo "[FOUT] Onbekende architectuur: $ARCH"; exit 1 ;;
esac

URL="https://github.com/$REPO/releases/download/$VERSION/$ZIPFILE"

echo "Bolttunnel $VERSION installeren..."
echo "  Architectuur : $ARCH"
echo "  Download     : $URL"
echo "  Installatie  : $INSTALL_DIR"
echo ""

# Controleer vereisten
for cmd in wget unzip; do
    if ! command -v $cmd &>/dev/null; then
        echo "[FOUT] '$cmd' niet gevonden. Installeer met: sudo apt install $cmd"
        exit 1
    fi
done

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

wget -q --show-progress "$URL" -O bolttunnel.zip
unzip -o bolttunnel.zip
rm bolttunnel.zip

chmod +x bolttunnel-client-linux-* bolttunnel-server-linux-* 2>/dev/null || true
chmod +x start-client.sh start-server.sh 2>/dev/null || true

echo ""
echo "Installatie klaar in $INSTALL_DIR"
echo ""
echo "Volgende stap:"
echo "  Pas start-client.sh aan met jouw server hostname"
echo "  Dan: ./start-client.sh"
