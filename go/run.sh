#!/bin/bash
# run.sh — Kaynak koddan doğrudan çalıştırma (Linux / macOS)
# Geliştirme sırasında kullanın. Release için derle.sh'ı kullanın.

set -e
cd "$(dirname "$0")"
echo "Çalıştırılıyor..."
go run .
