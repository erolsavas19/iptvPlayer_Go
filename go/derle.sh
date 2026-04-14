#!/bin/bash
# derle.sh — Linux / macOS için release build scripti
#
# Gereksinimler (Linux):
#   sudo apt install gcc libgl1-mesa-dev xorg-dev
#
# Gereksinimler (macOS):
#   xcode-select --install
#   brew install --cask vlc   (VLC için)
#
# Kullanım:
#   chmod +x derle.sh && ./derle.sh

set -e
cd "$(dirname "$0")"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Mimariyi Go formatına çevir (arm64 = Apple Silicon / aarch64)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

OUTPUT="../iptvPlayer_go_${OS}_${ARCH}"

echo "=== IPTV Player — Release Build ==="
echo "Platform : $OS/$ARCH"
echo "Çıktı    : $OUTPUT"
echo ""

go build \
    -ldflags="-s -w" \
    -o "$OUTPUT" \
    .

echo ""
echo "Build başarılı! Çalıştırmak için:"
echo "  $OUTPUT"
echo ""
echo "VLC kurulu değilse:"
if [ "$OS" = "linux" ]; then
    echo "  sudo apt install vlc"
elif [ "$OS" = "darwin" ]; then
    echo "  brew install --cask vlc"
fi
