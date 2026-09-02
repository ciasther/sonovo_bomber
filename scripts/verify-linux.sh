#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
rm -rf evidence
mkdir -p dist evidence
go test ./...
go vet ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
  -trimpath \
  -buildvcs=false \
  -ldflags='-s -w -buildid= -H=windowsgui' \
  -o dist/BomberRush.exe \
  ./cmd/bomberrush
go run ./cmd/rendercheck -root "$PWD" -out "$PWD/evidence"
